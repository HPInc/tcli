// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline_test

import (
	"strings"
	"testing"

	"github.com/hpinc/tcli/pkg/pipeline"
)

// loginUser answers plain text, logoutUser answers 200 with no body.
const textResponseYAML = `
name: integration-text
defaults:
  server: %s
  scheme: http
  retryCount: 0
steps:
  - name: login
    command: petstore user loginUser
    params:
      username: tester
      password: secret
    outputs:
      session: 'if (.text | startswith("logged in user session:")) then .text else error("unexpected login reply") end'
  - name: logout
    command: petstore user logoutUser
    dependsOn: [login]
`

func TestIntegration_TextAndBlankResponses(t *testing.T) {
	integrationEnabled(t)
	state := runPipelineAgainst(t, textResponseYAML)
	assertSucceeded(t, state, "login", "logout")

	if got, want := state.Get("login").Outputs["session"], "logged in user session:1234"; got != want {
		t.Errorf("login session = %v, want %q", got, want)
	}
	if n := len(state.Get("logout").Records); n != 0 {
		t.Errorf("a blank response should give no record, got %d", n)
	}
}

// A check on a plain-text response must still be able to fail the step.
func TestIntegration_TextResponseCheckFails(t *testing.T) {
	integrationEnabled(t)
	yaml := strings.Replace(textResponseYAML, `startswith("logged in user session:")`, `startswith("something else")`, 1)
	state, err := tryPipelineAgainst(t, yaml)
	if err == nil {
		t.Fatal("expected the failing outputs check to fail the pipeline")
	}
	login := state.Get("login")
	if login.Status != pipeline.StatusFailed || !strings.Contains(login.Err.Error(), "unexpected login reply") {
		t.Errorf("login: status=%s, err=%v; want an outputs failure", login.Status, login.Err)
	}
}
