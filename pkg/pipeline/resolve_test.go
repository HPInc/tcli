// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"strings"
	"testing"
)

func newTestState() *State {
	p := &Pipeline{
		Variables: map[string]any{
			"petCount": 5,
			"status":   404,
			"nested":   map[string]any{"key": "value"},
		},
		Steps: []*Step{
			{Name: "a", Command: "m o"},
			{Name: "b", Command: "m o"},
		},
	}
	s := NewState(p)
	s.Set("a", &StepResult{
		Status:  StatusSucceeded,
		Outputs: map[string]any{"id": float64(42), "info": map[string]any{"tag": "prod"}},
	})
	s.Set("b", &StepResult{Status: StatusFailed})
	return s
}

func TestResolveString_Variables(t *testing.T) {
	r := NewResolver(newTestState())
	got, err := r.ResolveString("count=${{ variables.petCount }}")
	if err != nil {
		t.Fatalf("ResolveString: %v", err)
	}
	if got != "count=5" {
		t.Errorf("got %q, want count=5", got)
	}
}

func TestResolveString_StepOutput(t *testing.T) {
	r := NewResolver(newTestState())
	got, err := r.ResolveString("id=${{ steps.a.outputs.id }}")
	if err != nil {
		t.Fatalf("ResolveString: %v", err)
	}
	if got != "id=42" {
		t.Errorf("got %q, want id=42", got)
	}
}

func TestResolveValue_TypedForSingleRef(t *testing.T) {
	r := NewResolver(newTestState())
	got, err := r.ResolveValue("${{ steps.a.ok }}")
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	if got != true {
		t.Errorf("got %v (%T), want bool true", got, got)
	}
}

func TestResolveValue_StringInterpolationForMixed(t *testing.T) {
	r := NewResolver(newTestState())
	got, err := r.ResolveValue("prefix-${{ variables.petCount }}-suffix")
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	if got != "prefix-5-suffix" {
		t.Errorf("got %v, want prefix-5-suffix", got)
	}
}

func TestResolveAny_NestedMap(t *testing.T) {
	r := NewResolver(newTestState())
	in := map[string]any{
		"petId":   "${{ steps.a.outputs.id }}",
		"tag":     "${{ steps.a.outputs.info.tag }}",
		"literal": 42,
		"deep": map[string]any{
			"x": "${{ variables.petCount }}",
		},
	}
	out, err := r.ResolveAny(in)
	if err != nil {
		t.Fatalf("ResolveAny: %v", err)
	}
	m := out.(map[string]any)
	if m["petId"] != float64(42) {
		t.Errorf("petId = %v, want 42", m["petId"])
	}
	if m["tag"] != "prod" {
		t.Errorf("tag = %v, want prod", m["tag"])
	}
	if m["literal"] != 42 {
		t.Errorf("literal = %v, want 42", m["literal"])
	}
	if m["deep"].(map[string]any)["x"] != 5 {
		t.Errorf("deep.x = %v, want 5", m["deep"].(map[string]any)["x"])
	}
}

func TestResolve_UnknownVariable(t *testing.T) {
	r := NewResolver(newTestState())
	_, err := r.ResolveString("${{ variables.ghost }}")
	if err == nil || !strings.Contains(err.Error(), "unknown variable") {
		t.Fatalf("expected unknown-variable error, got %v", err)
	}
}

func TestResolve_UnknownStep(t *testing.T) {
	r := NewResolver(newTestState())
	_, err := r.ResolveString("${{ steps.ghost.ok }}")
	if err == nil || !strings.Contains(err.Error(), "unknown step") {
		t.Fatalf("expected unknown-step error, got %v", err)
	}
}

func TestResolve_EnvHook(t *testing.T) {
	s := newTestState()
	r := &Resolver{State: s, Env: func(k string) string {
		if k == "TCLI_TEST" {
			return "hello"
		}
		return ""
	}}
	got, err := r.ResolveString("${{ env.TCLI_TEST }}")
	if err != nil || got != "hello" {
		t.Errorf("got %q, err=%v; want hello", got, err)
	}
}
