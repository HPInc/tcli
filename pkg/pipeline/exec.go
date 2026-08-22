// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import "context"

// Record is one JSON object flowing between steps.
type Record = map[string]any

// Runner is the boundary between the pipeline engine and the tcli command
// execution layer. Production code uses SubprocessRunner (which shells out to
// the tcli binary); tests substitute an in-memory fake.
//
// A Runner receives the fully-resolved step (interpolation done, defaults
// merged) plus any records streamed in from an upstream step, and returns the
// records this step produced.
type Runner interface {
	Run(ctx context.Context, s *Step, input []Record) ([]Record, error)
}

// Executor walks the pipeline DAG, resolves interpolations, and dispatches
// each step to the configured Runner. Zero value is not usable; use
// NewExecutor.
type Executor struct {
	Runner Runner
}

// NewExecutor returns an Executor that dispatches steps to r.
func NewExecutor(r Runner) *Executor { return &Executor{Runner: r} }

// Run executes p to completion and returns the shared runtime state. If any
// non-continueOnError step fails, Run returns the first such error (all other
// steps are still reflected in the returned State). If the underlying DAG
// build fails, Run returns that error directly.
func (e *Executor) Run(ctx context.Context, p *Pipeline) (*State, error) {
	dag, err := BuildDAG(p)
	if err != nil {
		return nil, err
	}
	state := NewState(p)
	sched := newScheduler(p, dag, state, e.Runner)
	return state, sched.run(ctx)
}
