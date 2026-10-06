// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package env

import (
	"bytes"
	"errors"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	originalVersion := version
	t.Cleanup(func() {
		version = originalVersion
	})

	for _, value := range []string{"dev", "v1.2.3", "v1.2.3-rc.1", "v1.2.3+build.1"} {
		t.Run(value, func(t *testing.T) {
			version = value
			var output bytes.Buffer
			handled, err := printVersion([]string{"tcli", "version"}, &output)
			if err != nil {
				t.Fatalf("printVersion returned an error: %v", err)
			}
			if !handled {
				t.Fatal("printVersion did not handle the version command")
			}
			if got, want := output.String(), value+"\n"; got != want {
				t.Errorf("printVersion output = %q, want %q", got, want)
			}
		})
	}
}

func TestPrintVersionIgnoresOtherCommands(t *testing.T) {
	for _, args := range [][]string{nil, {"tcli"}, {"tcli", "petstore"}} {
		var output bytes.Buffer
		handled, err := printVersion(args, &output)
		if err != nil {
			t.Fatalf("printVersion returned an error: %v", err)
		}
		if handled {
			t.Fatalf("printVersion handled non-version arguments %q", args)
		}
		if output.Len() != 0 {
			t.Errorf("printVersion wrote %q for non-version arguments", output.String())
		}
	}
}

type failingVersionWriter struct {
	err error
}

func (w failingVersionWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestPrintVersionWriteError(t *testing.T) {
	want := errors.New("write failed")
	handled, err := printVersion([]string{"tcli", "version"}, failingVersionWriter{err: want})
	if !handled || !errors.Is(err, want) {
		t.Fatalf("printVersion = (%v, %v), want (true, %v)", handled, err, want)
	}
}
