// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package env

import (
	"bytes"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	originalVersion := version
	version = "v1.2.3"
	t.Cleanup(func() {
		version = originalVersion
	})

	var output bytes.Buffer
	handled, err := printVersion([]string{"tcli", "version"}, &output)
	if err != nil {
		t.Fatalf("printVersion returned an error: %v", err)
	}
	if !handled {
		t.Fatal("printVersion did not handle the version command")
	}
	if got, want := output.String(), "v1.2.3\n"; got != want {
		t.Errorf("printVersion output = %q, want %q", got, want)
	}
}

func TestPrintVersionIgnoresOtherCommands(t *testing.T) {
	var output bytes.Buffer
	handled, err := printVersion([]string{"tcli", "petstore"}, &output)
	if err != nil {
		t.Fatalf("printVersion returned an error: %v", err)
	}
	if handled {
		t.Fatal("printVersion handled a non-version command")
	}
	if output.Len() != 0 {
		t.Errorf("printVersion wrote %q for a non-version command", output.String())
	}
}

func TestPrintVersionNoArgs(t *testing.T) {
	var output bytes.Buffer
	handled, err := printVersion([]string{"tcli"}, &output)
	if err != nil {
		t.Fatalf("printVersion returned an error: %v", err)
	}
	if handled {
		t.Fatal("printVersion handled a bare invocation with no subcommand")
	}
}
