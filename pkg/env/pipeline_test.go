// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package env

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/hpinc/tcli/pkg/pipeline"
)

func TestRunPipelineFilesMixedResults(t *testing.T) {
	dir := t.TempDir()
	good := writeFile(t, filepath.Join(dir, "good.yaml"), pipelineYAML("good"))
	invalid := writeFile(t, filepath.Join(dir, "invalid.yaml"), "not: [valid")
	missing := filepath.Join(dir, "missing.yaml")
	boom := writeFile(t, filepath.Join(dir, "boom.yaml"), pipelineYAML("boom"))
	last := writeFile(t, filepath.Join(dir, "last.yaml"), pipelineYAML("last"))

	var ran []string
	err := runPipelineFiles([]string{good, invalid, missing, boom, last}, false, func(p *pipeline.Pipeline) (*pipeline.State, error) {
		ran = append(ran, p.Name)
		if p.Name == "boom" {
			return nil, errors.New("execution failed")
		}
		return nil, nil
	})

	if want := []string{"good", "boom", "last"}; !slices.Equal(ran, want) {
		t.Fatalf("pipelines ran %v, want %v", ran, want)
	}
	for _, want := range []string{invalid, missing, "execution failed"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %v, want it to contain %q", err, want)
		}
	}
}

func TestRunPipelineFilesSingleFileKeepsErrorText(t *testing.T) {
	dir := t.TempDir()
	good := writeFile(t, filepath.Join(dir, "good.yaml"), pipelineYAML("good"))
	invalid := writeFile(t, filepath.Join(dir, "invalid.yaml"), "not: [valid")
	execErr := errors.New("execution failed")
	ok := func(*pipeline.Pipeline) (*pipeline.State, error) { return nil, nil }

	if err := runPipelineFiles([]string{good}, false, ok); err != nil {
		t.Fatalf("valid pipeline returned %v", err)
	}
	err := runPipelineFiles([]string{good}, false, func(*pipeline.Pipeline) (*pipeline.State, error) {
		return nil, execErr
	})
	if err == nil || err.Error() != execErr.Error() {
		t.Fatalf("error = %v, want %q", err, execErr)
	}
	err = runPipelineFiles([]string{invalid}, false, ok)
	if err == nil || strings.HasPrefix(err.Error(), invalid+": ") {
		t.Fatalf("error = %v, want unprefixed load error", err)
	}
}

func pipelineYAML(name string) string {
	return "name: " + name + "\nsteps:\n  - name: step\n    command: module operation\n"
}

func writeFile(t *testing.T, path, content string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
