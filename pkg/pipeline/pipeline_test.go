// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ExampleCRUD(t *testing.T) {
	p, err := Load(filepath.Join("..", "..", "examples", "pipeline", "petstore_crud.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p.Name != "petstore-crud" {
		t.Errorf("name = %q, want petstore-crud", p.Name)
	}
	if got, want := len(p.Steps), 5; got != want {
		t.Fatalf("steps = %d, want %d", got, want)
	}

	dag, err := BuildDAG(p)
	if err != nil {
		t.Fatalf("BuildDAG: %v", err)
	}
	// Linear pipeline: order must match declaration.
	wantOrder := []string{"seed", "create", "read", "delete", "verify_gone"}
	for i, n := range dag.Nodes {
		if n.Step.Name != wantOrder[i] {
			t.Errorf("topo[%d] = %q, want %q", i, n.Step.Name, wantOrder[i])
		}
	}
}

func TestLoad_ExampleDAG(t *testing.T) {
	p, err := Load(filepath.Join("..", "..", "examples", "pipeline", "petstore_dag.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	dag, err := BuildDAG(p)
	if err != nil {
		t.Fatalf("BuildDAG: %v", err)
	}

	// create_one must precede everything that references it.
	pos := stepPositions(dag)
	for _, later := range []string{"read_by_id", "search_by_status", "delete_one", "verify_gone"} {
		if pos["create_one"] >= pos[later] {
			t.Errorf("create_one (%d) should come before %s (%d)",
				pos["create_one"], later, pos[later])
		}
	}
	// delete_one must wait on both branches (fan-in).
	if pos["delete_one"] <= pos["read_by_id"] || pos["delete_one"] <= pos["search_by_status"] {
		t.Errorf("delete_one should come after both read_by_id and search_by_status")
	}
}

func TestLoad_ExampleParallel(t *testing.T) {
	p, err := Load(filepath.Join("..", "..", "examples", "pipeline", "petstore_parallel.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p.Concurrency == nil || p.Concurrency.MaxParallel != 8 {
		t.Errorf("expected concurrency.maxParallel=8, got %+v", p.Concurrency)
	}
	if _, err := BuildDAG(p); err != nil {
		t.Fatalf("BuildDAG: %v", err)
	}
}

func TestValidate_RejectsReservedParams(t *testing.T) {
	yaml := `
name: bad
steps:
  - name: a
    command: petstore pet addPet
    params:
      status_code: 200
`
	_, err := Parse([]byte(yaml))
	if err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("expected reserved-key error, got %v", err)
	}
}

func TestValidate_RejectsCycle(t *testing.T) {
	yaml := `
name: cyclic
steps:
  - name: a
    command: m op
    dependsOn: [b]
  - name: b
    command: m op
    dependsOn: [a]
`
	_, err := Parse([]byte(yaml))
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

func TestValidate_RejectsUnknownStepRef(t *testing.T) {
	yaml := `
name: bad
steps:
  - name: a
    command: m op
    inputFrom: ghost
`
	_, err := Parse([]byte(yaml))
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("expected unknown-step error, got %v", err)
	}
}

func TestValidate_RejectsExpressionRefToUnknownStep(t *testing.T) {
	yaml := `
name: bad
steps:
  - name: a
    command: m op
    params:
      petId: "${{ steps.ghost.outputs.id }}"
`
	_, err := Parse([]byte(yaml))
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("expected unknown-step error, got %v", err)
	}
}

func TestFindReferences(t *testing.T) {
	got := FindReferences("hello ${{ variables.x }} and ${{ steps.a.outputs.b }}")
	if len(got) != 2 {
		t.Fatalf("got %d refs, want 2", len(got))
	}
	if got[0].Kind() != "variables" || got[0].Path[1] != "x" {
		t.Errorf("ref[0] = %+v", got[0])
	}
	if got[1].StepDep() != "a" {
		t.Errorf("ref[1].StepDep = %q, want a", got[1].StepDep())
	}
}

func stepPositions(d *DAG) map[string]int {
	pos := make(map[string]int, len(d.Nodes))
	for i, n := range d.Nodes {
		pos[n.Step.Name] = i
	}
	return pos
}
