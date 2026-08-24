// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hpinc/tcli/internal/petstoretest"
	"github.com/hpinc/tcli/pkg/pipeline"
)

// Integration tests exercise the full stack: pipeline YAML -> Executor ->
// SubprocessRunner -> tcli binary -> local test-server HTTP -> assertions
// on the pipeline's runtime state.
//
// Gated on TCLI_INTEGRATION so the default `go test` stays fast. CI should
// set the env var to run them.

func integrationEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("TCLI_INTEGRATION") == "" {
		t.Skip("set TCLI_INTEGRATION=1 to run integration tests")
	}
}

// buildOnce compiles the tcli binary once per test package invocation.
// The subprocess runner needs an actual executable go run would be
// simpler but re-compiles on every step, which is far too slow.
var (
	buildOnce sync.Once
	buildBin  string
	buildErr  error
)

func tcliBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "tcli-integration-*")
		if err != nil {
			buildErr = err
			return
		}
		buildBin = filepath.Join(dir, "tcli")
		cwd, _ := os.Getwd()
		repoRoot := filepath.Join(cwd, "..", "..")
		cmd := exec.Command("go", "build", "-o", buildBin, "./cmd")
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("build tcli: %w: %s", err, out)
		}
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return buildBin
}

// runPipelineAgainst spins up the fake petstore, points the given YAML
// template at it, then runs it via SubprocessRunner and returns final state.
// yamlTemplate must contain one %s where the server host:port goes.
func runPipelineAgainst(t *testing.T, yamlTemplate string) *pipeline.State {
	t.Helper()

	binary := tcliBinary(t)
	server := petstoretest.NewServer()
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	// Point the subprocess at the source-tree tools/ so it finds petstore.json.
	cwd, _ := os.Getwd()
	repoRoot := filepath.Join(cwd, "..", "..")
	t.Setenv("TCLI_CONFIG_ROOT", filepath.Join(repoRoot, "tools"))

	yaml := fmt.Sprintf(yamlTemplate, u.Host)
	file := filepath.Join(t.TempDir(), "pipeline.yaml")
	if err := os.WriteFile(file, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}

	p, err := pipeline.Load(file)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	runner := &pipeline.SubprocessRunner{Binary: binary}
	state, err := pipeline.NewExecutor(runner).Run(context.Background(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return state
}

func assertSucceeded(t *testing.T, state *pipeline.State, names ...string) {
	t.Helper()
	for _, name := range names {
		r := state.Get(name)
		if r == nil {
			t.Errorf("step %q: no result recorded", name)
			continue
		}
		if r.Status != pipeline.StatusSucceeded {
			t.Errorf("step %q: status=%s, err=%v", name, r.Status, r.Err)
		}
	}
}

const crudYAML = `
name: integration-crud
defaults:
  server: %s
  scheme: http
variables:
  petCount: 3
  status_notfound: 404
steps:
  - name: seed
    command: utils echo
    format: 'range(1;${{ variables.petCount }}+1) | {body: {id:., name:.|tostring, photourls:[.|tostring]}}'
  - name: create
    command: petstore pet addPet
    inputFrom: seed
    format: '{petId:.id}'
  - name: read
    command: petstore pet getPetById
    inputFrom: create
    format: '{petId:.id, api_key:.id}'
  - name: delete
    command: petstore pet deletePet
    inputFrom: read
    format: '{petId:.message}'
  - name: verify_gone
    command: petstore pet getPetById
    inputFrom: delete
    statusCode: ${{ variables.status_notfound }}
`

func TestIntegration_PetstoreCrud(t *testing.T) {
	integrationEnabled(t)
	state := runPipelineAgainst(t, crudYAML)
	assertSucceeded(t, state, "seed", "create", "read", "delete", "verify_gone")
	if got := len(state.Get("create").Records); got != 3 {
		t.Errorf("create produced %d records, want 3", got)
	}
}

const dagYAML = `
name: integration-dag
defaults:
  server: %s
  scheme: http
variables:
  petId: 4242
  status_notfound: 404
steps:
  - name: create_one
    command: petstore pet addPet
    body:
      id: ${{ variables.petId }}
      name: fido
      photoUrls: ["http://img.example/fido.png"]
    outputs:
      newId: .id
  - name: read_by_id
    command: petstore pet getPetById
    params:
      petId: ${{ steps.create_one.outputs.newId }}
  - name: search_by_status
    command: petstore pet findPetsByStatus
    params:
      status: available
    dependsOn: [create_one]
  - name: delete_one
    command: petstore pet deletePet
    params:
      petId: ${{ steps.create_one.outputs.newId }}
    dependsOn: [read_by_id, search_by_status]
  - name: verify_gone
    command: petstore pet getPetById
    params:
      petId: ${{ steps.create_one.outputs.newId }}
    dependsOn: [delete_one]
    statusCode: ${{ variables.status_notfound }}
`

func TestIntegration_PetstoreDag(t *testing.T) {
	integrationEnabled(t)
	state := runPipelineAgainst(t, dagYAML)
	assertSucceeded(t, state,
		"create_one", "read_by_id", "search_by_status", "delete_one", "verify_gone")

	// Scalar-reference resolution: created id (from response) should be 4242
	// and delete_one used that value to remove the correct pet proven by
	// verify_gone returning 404.
	newId := state.Get("create_one").Outputs["newId"]
	if newId != float64(4242) {
		t.Errorf("create_one.outputs.newId = %v, want 4242", newId)
	}
}
