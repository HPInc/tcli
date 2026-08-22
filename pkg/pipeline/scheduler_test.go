// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeRunner records every invocation and returns pre-canned responses
// keyed by step name. Missing entries return an empty record set.
type fakeRunner struct {
	mu          sync.Mutex
	calls       []call
	responses   map[string][]Record
	errs        map[string]error
	delays      map[string]time.Duration
	inFlight    int32
	maxInFlight int32
}

type call struct {
	step  string
	input []Record
}

func (f *fakeRunner) Run(ctx context.Context, s *Step, input []Record) ([]Record, error) {
	cur := atomic.AddInt32(&f.inFlight, 1)
	for {
		old := atomic.LoadInt32(&f.maxInFlight)
		if cur <= old || atomic.CompareAndSwapInt32(&f.maxInFlight, old, cur) {
			break
		}
	}
	defer atomic.AddInt32(&f.inFlight, -1)

	if d, ok := f.delays[s.Name]; ok {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	f.mu.Lock()
	f.calls = append(f.calls, call{step: s.Name, input: input})
	f.mu.Unlock()

	if err, ok := f.errs[s.Name]; ok {
		return nil, err
	}
	return f.responses[s.Name], nil
}

func (f *fakeRunner) callOrder() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	for i, c := range f.calls {
		out[i] = c.step
	}
	return out
}

func TestExecutor_LinearPipeThroughInputFrom(t *testing.T) {
	yaml := `
name: linear
steps:
  - name: a
    command: m op
  - name: b
    command: m op
    inputFrom: a
`
	p, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	runner := &fakeRunner{
		responses: map[string][]Record{
			"a": {{"id": float64(1)}, {"id": float64(2)}},
		},
	}
	state, err := NewExecutor(runner).Run(context.Background(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := runner.callOrder(), []string{"a", "b"}; !equalStrings(got, want) {
		t.Errorf("call order = %v, want %v", got, want)
	}
	// b received a's two records as input
	if len(runner.calls[1].input) != 2 {
		t.Errorf("b input len = %d, want 2", len(runner.calls[1].input))
	}
	if state.Get("a").Status != StatusSucceeded || state.Get("b").Status != StatusSucceeded {
		t.Errorf("expected both succeeded, got %+v", state.Snapshot())
	}
}

func TestExecutor_FanOutAndFanIn(t *testing.T) {
	yaml := `
name: dag
steps:
  - name: root
    command: m op
  - name: left
    command: m op
    dependsOn: [root]
  - name: right
    command: m op
    dependsOn: [root]
  - name: join
    command: m op
    dependsOn: [left, right]
`
	p, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	runner := &fakeRunner{}
	_, err = NewExecutor(runner).Run(context.Background(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	order := runner.callOrder()
	pos := map[string]int{}
	for i, name := range order {
		pos[name] = i
	}
	if pos["root"] != 0 {
		t.Errorf("root should run first, got position %d", pos["root"])
	}
	if pos["join"] <= pos["left"] || pos["join"] <= pos["right"] {
		t.Errorf("join should run after left and right, order=%v", order)
	}
}

func TestExecutor_FailureCancelsSiblings(t *testing.T) {
	yaml := `
name: fail
concurrency:
  onFailure: cancel
steps:
  - name: root
    command: m op
  - name: left
    command: m op
    dependsOn: [root]
  - name: right
    command: m op
    dependsOn: [root]
  - name: after_left
    command: m op
    dependsOn: [left]
`
	p, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	runner := &fakeRunner{
		errs:   map[string]error{"left": errors.New("boom")},
		delays: map[string]time.Duration{"right": 100 * time.Millisecond},
	}
	state, runErr := NewExecutor(runner).Run(context.Background(), p)
	if runErr == nil || !strings.Contains(runErr.Error(), "boom") {
		t.Fatalf("expected boom error, got %v", runErr)
	}
	// after_left must be skipped (parent failed)
	if state.Get("after_left").Status != StatusSkipped {
		t.Errorf("after_left status = %s, want skipped", state.Get("after_left").Status)
	}
}

func TestExecutor_ContinueOnErrorRunsDownstreamOfSameStep(t *testing.T) {
	// continueOnError makes the *pipeline* survive; downstream of a failed
	// step still skips (because parentsOk returns false).
	yaml := `
name: cont
steps:
  - name: a
    command: m op
    continueOnError: true
  - name: b
    command: m op
    dependsOn: [a]
`
	p, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	runner := &fakeRunner{errs: map[string]error{"a": errors.New("boom")}}
	state, runErr := NewExecutor(runner).Run(context.Background(), p)
	if runErr != nil {
		t.Fatalf("continueOnError should hide the error, got %v", runErr)
	}
	if state.Get("a").Status != StatusFailed {
		t.Errorf("a status = %s, want failed", state.Get("a").Status)
	}
	if state.Get("b").Status != StatusSkipped {
		t.Errorf("b status = %s, want skipped", state.Get("b").Status)
	}
}

func TestExecutor_ConditionSkipsStep(t *testing.T) {
	yaml := `
name: cond
steps:
  - name: a
    command: m op
    continueOnError: true
  - name: b
    command: m op
    condition: "${{ steps.a.ok }}"
`
	p, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	runner := &fakeRunner{errs: map[string]error{"a": errors.New("boom")}}
	state, _ := NewExecutor(runner).Run(context.Background(), p)
	if state.Get("b").Status != StatusSkipped {
		t.Errorf("b status = %s, want skipped (condition false)", state.Get("b").Status)
	}
}

func TestExecutor_MaxParallelCaps(t *testing.T) {
	yaml := `
name: cap
concurrency:
  maxParallel: 2
steps:
  - name: a
    command: m op
  - name: b
    command: m op
  - name: c
    command: m op
  - name: d
    command: m op
`
	p, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	delays := map[string]time.Duration{
		"a": 50 * time.Millisecond, "b": 50 * time.Millisecond,
		"c": 50 * time.Millisecond, "d": 50 * time.Millisecond,
	}
	runner := &fakeRunner{delays: delays}
	_, err = NewExecutor(runner).Run(context.Background(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if runner.maxInFlight > 2 {
		t.Errorf("maxInFlight = %d, want <= 2", runner.maxInFlight)
	}
}

func TestExecutor_OutputsCapturedFromLastRecord(t *testing.T) {
	yaml := `
name: out
steps:
  - name: a
    command: m op
    outputs:
      id: .id
      name: .name
`
	p, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	runner := &fakeRunner{
		responses: map[string][]Record{
			"a": {
				{"id": float64(1), "name": "first"},
				{"id": float64(2), "name": "last"},
			},
		},
	}
	state, err := NewExecutor(runner).Run(context.Background(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	outs := state.Get("a").Outputs
	if outs["id"] != float64(2) {
		t.Errorf("outputs.id = %v, want 2 (last record)", outs["id"])
	}
	if outs["name"] != "last" {
		t.Errorf("outputs.name = %v, want last", outs["name"])
	}
}

func TestBuildArgs_Mapping(t *testing.T) {
	rc := 3
	pll := 8
	verb := true
	s := &Step{
		Name:        "x",
		Command:     "petstore pet addPet",
		Format:      "{petId:.id}",
		StatusCode:  "200",
		Count:       5,
		RetryCount:  &rc,
		Parallelism: &pll,
		Verbose:     &verb,
		Params:      map[string]any{"petId": 42},
		Body:        map[string]any{"id": 42, "name": "fido"},
	}
	args, err := BuildArgs(s)
	if err != nil {
		t.Fatalf("BuildArgs: %v", err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{
		"petstore pet addPet",
		"-format {petId:.id}",
		"-status_code 200",
		"-count 5",
		"-retry_count 3",
		"-v",
		"-parallel",
		"-body ",
		`"id":42`,
		"-petId 42",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("args missing %q; got %q", want, joined)
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
