// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package env

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hpinc/tcli/pkg/pipeline"
)

func TestRunPipelineFilesRunsAllFilesInOrder(t *testing.T) {
	dir := t.TempDir()
	first := writePipelineFile(t, dir, "first.yaml", "first")
	invalid := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(invalid, []byte("not: [valid"), 0600); err != nil {
		t.Fatal(err)
	}
	last := writePipelineFile(t, dir, "last.yaml", "last")

	var ran []string
	err := runPipelineFiles([]string{first, invalid, last}, false, func(p *pipeline.Pipeline) (*pipeline.State, error) {
		ran = append(ran, p.Name)
		if p.Name == "first" {
			return nil, errors.New("execution failed")
		}
		return nil, nil
	})

	if strings.Join(ran, ",") != "first,last" {
		t.Fatalf("pipelines ran in order %v, want [first last]", ran)
	}
	if err == nil {
		t.Fatal("expected errors from the failed pipeline and invalid YAML")
	}
	if !strings.Contains(err.Error(), "execution failed") || !strings.Contains(err.Error(), invalid) {
		t.Fatalf("joined error %q does not include both failures", err)
	}
}

func writePipelineFile(t *testing.T, dir, filename, name string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	content := "name: " + name + "\nsteps:\n  - name: step\n    command: module operation\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
