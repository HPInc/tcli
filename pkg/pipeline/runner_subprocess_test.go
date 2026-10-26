// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"context"
	"strings"
	"testing"
)

func TestSubprocessRunner_RedactsJWTFromError(t *testing.T) {
	jwt := "sensitive-token"
	runner := &SubprocessRunner{Binary: "/usr/bin/false"}

	_, err := runner.Run(context.Background(), &Step{
		Name:    "failure",
		Command: "demo operation",
		Jwt:     &jwt,
	}, nil)
	if err == nil {
		t.Fatal("expected subprocess error")
	}
	if strings.Contains(err.Error(), jwt) {
		t.Fatalf("error exposed JWT: %v", err)
	}
	if !strings.Contains(err.Error(), "-jwt ******") {
		t.Errorf("error does not show redacted JWT flag: %v", err)
	}
}

func TestRedactSecretArgs(t *testing.T) {
	args := []string{"demo", "-jwt", "token", "-password", "secret"}
	redacted := redactSecretArgs(args, map[string]struct{}{
		"-jwt":      {},
		"-password": {},
	})
	if got, want := strings.Join(redacted, " "), "demo -jwt ****** -password ******"; got != want {
		t.Errorf("redacted args = %q, want %q", got, want)
	}
	if got, want := strings.Join(args, " "), "demo -jwt token -password secret"; got != want {
		t.Errorf("original args mutated to %q, want %q", got, want)
	}
}
