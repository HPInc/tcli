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

func TestRunPipelineFilesSingleFileKeepsErrorText(t *testing.T) {
	dir := t.TempDir()
	invalid := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(invalid, []byte("not: [valid"), 0600); err != nil {
		t.Fatal(err)
	}
	_, loadErr := pipeline.Load(invalid)

	noExec := func(*pipeline.Pipeline) (*pipeline.State, error) {
		t.Fatal("execute must not run for a file that fails to load")
		return nil, nil
	}
	err := runPipelineFiles([]string{invalid}, false, noExec)
	if err == nil || err.Error() != loadErr.Error() {
		t.Fatalf("single-file load error = %v, want %v", err, loadErr)
	}

	execErr := errors.New("execution failed")
	good := writePipelineFile(t, dir, "good.yaml", "good")
	err = runPipelineFiles([]string{good}, false, func(*pipeline.Pipeline) (*pipeline.State, error) {
		return nil, execErr
	})
	if err == nil || err.Error() != execErr.Error() {
		t.Fatalf("single-file execution error = %v, want %v", err, execErr)
	}
}

func TestRunPipelineFilesMultipleSucceed(t *testing.T) {
	dir := t.TempDir()
	a := writePipelineFile(t, dir, "a.yaml", "a")
	b := writePipelineFile(t, dir, "b.yaml", "b")

	var ran []string
	err := runPipelineFiles([]string{a, b}, false, func(p *pipeline.Pipeline) (*pipeline.State, error) {
		ran = append(ran, p.Name)
		return nil, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(ran, ",") != "a,b" {
		t.Fatalf("pipelines ran in order %v, want [a b]", ran)
	}
}

func TestRunPipelineFilesMissingFileNamesPath(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.yaml")
	ok := writePipelineFile(t, dir, "ok.yaml", "ok")

	var ran []string
	err := runPipelineFiles([]string{missing, ok}, false, func(p *pipeline.Pipeline) (*pipeline.State, error) {
		ran = append(ran, p.Name)
		return nil, nil
	})
	if err == nil || !strings.Contains(err.Error(), missing) {
		t.Fatalf("error = %v, want it to name %s", err, missing)
	}
	if strings.Join(ran, ",") != "ok" {
		t.Fatalf("pipelines ran %v, want [ok]", ran)
	}
}

func TestRunPipelineRunRequiresFile(t *testing.T) {
	err := runPipelineRun(nil)
	if err == nil || !strings.Contains(err.Error(), "usage: tcli pipeline run") {
		t.Fatalf("error = %v, want usage message", err)
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
