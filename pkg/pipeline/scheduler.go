// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/itchyny/gojq"
)

// scheduler walks the DAG and drives the Runner. It owns concurrency
// (maxParallel), failure policy (onFailure), and the "skip downstream when
// parents did not succeed" rule.
type scheduler struct {
	p      *Pipeline
	dag    *DAG
	state  *State
	runner Runner
	sem    chan struct{} // maxParallel cap; nil = unlimited
	onFail string

	mu       sync.Mutex
	firstErr error
	cancel   context.CancelFunc
}

func newScheduler(p *Pipeline, dag *DAG, state *State, runner Runner) *scheduler {
	s := &scheduler{p: p, dag: dag, state: state, runner: runner, onFail: OnFailureCancel}
	if p.Concurrency != nil {
		if p.Concurrency.MaxParallel > 0 {
			s.sem = make(chan struct{}, p.Concurrency.MaxParallel)
		}
		if p.Concurrency.OnFailure != "" {
			s.onFail = p.Concurrency.OnFailure
		}
	}
	return s
}

// run schedules every node and blocks until all nodes finish or the run is
// cancelled by a failure. Returns the first non-continueOnError error.
func (s *scheduler) run(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)
	defer s.cancel()

	var wg sync.WaitGroup
	done := make(chan *Node, len(s.dag.Nodes))

	remaining := make(map[*Node]int, len(s.dag.Nodes))
	for _, n := range s.dag.Nodes {
		remaining[n] = len(n.Deps)
		if remaining[n] == 0 {
			wg.Add(1)
			go s.execute(ctx, n, done, &wg)
		}
	}

	for finished := 0; finished < len(s.dag.Nodes); finished++ {
		n := <-done
		for _, child := range n.Dependents {
			remaining[child]--
			if remaining[child] == 0 {
				wg.Add(1)
				go s.execute(ctx, child, done, &wg)
			}
		}
	}
	wg.Wait()
	return s.firstErr
}

// execute drives one node: acquire the concurrency slot, check parent
// results, evaluate the condition, resolve interpolation, invoke the runner,
// extract outputs, and record the final StepResult.
func (s *scheduler) execute(ctx context.Context, n *Node, done chan<- *Node, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() { done <- n }()

	if s.sem != nil {
		select {
		case s.sem <- struct{}{}:
			defer func() { <-s.sem }()
		case <-ctx.Done():
			s.state.Set(n.Step.Name, &StepResult{Status: StatusSkipped})
			return
		}
	}

	if ctx.Err() != nil {
		s.state.Set(n.Step.Name, &StepResult{Status: StatusSkipped})
		return
	}

	if !s.parentsOk(n) {
		s.state.Set(n.Step.Name, &StepResult{Status: StatusSkipped})
		return
	}

	resolver := NewResolver(s.state)
	resolved, err := s.resolveStep(n.Step, resolver)
	if err != nil {
		s.finish(n, &StepResult{Status: StatusFailed, Err: fmt.Errorf("resolve: %w", err)})
		return
	}

	if skip, err := s.shouldSkip(resolved.Condition, resolver); err != nil {
		s.finish(n, &StepResult{Status: StatusFailed, Err: fmt.Errorf("condition: %w", err)})
		return
	} else if skip {
		s.state.Set(n.Step.Name, &StepResult{Status: StatusSkipped})
		return
	}

	input := s.gatherInput(n)
	records, err := s.runner.Run(ctx, resolved, input)
	result := &StepResult{Records: records}
	if err != nil {
		result.Status = StatusFailed
		result.Err = err
	} else {
		result.Status = StatusSucceeded
	}
	if outputs, oerr := extractOutputs(records, resolved.Outputs); oerr != nil && result.Err == nil {
		result.Status = StatusFailed
		result.Err = fmt.Errorf("outputs: %w", oerr)
	} else {
		result.Outputs = outputs
	}
	s.finish(n, result)
}

// finish records the result and, if this was a hard failure, remembers the
// first error and (per onFailure=cancel) tears down the run.
func (s *scheduler) finish(n *Node, r *StepResult) {
	s.state.Set(n.Step.Name, r)
	if r.Status != StatusFailed || n.Step.ContinueOnError {
		return
	}
	s.mu.Lock()
	if s.firstErr == nil {
		s.firstErr = fmt.Errorf("step %q failed: %w", n.Step.Name, r.Err)
	}
	s.mu.Unlock()
	if s.onFail == OnFailureCancel && s.cancel != nil {
		s.cancel()
	}
}

// parentsOk returns true only when every dep step succeeded. A dep failure
// (or a dep-of-a-dep-failure that skipped) propagates as skipped downstream.
// A dep that itself was continueOnError-failed still counts as "not ok" for
// downstream — matches ADO's "step failed even if the job continues".
func (s *scheduler) parentsOk(n *Node) bool {
	for _, dep := range n.Deps {
		if !s.state.Get(dep.Step.Name).Ok() {
			return false
		}
	}
	return true
}

// shouldSkip evaluates a step's condition. Empty condition -> don't skip.
// A condition that resolves to nil / false / 0 / "" -> skip.
func (s *scheduler) shouldSkip(cond string, r *Resolver) (bool, error) {
	if cond == "" {
		return false, nil
	}
	val, err := r.ResolveValue(cond)
	if err != nil {
		return false, err
	}
	return !truthy(val), nil
}

// gatherInput returns the records the step reads via inputFrom. Buffered:
// v1 always waits for the producer to finish before starting the consumer.
// True concurrent streaming (both endpoints running at once) is a v2 item.
func (s *scheduler) gatherInput(n *Node) []Record {
	if n.Step.InputFrom == "" {
		return nil
	}
	upstream := s.state.Get(n.Step.InputFrom)
	if upstream == nil {
		return nil
	}
	return upstream.Records
}

// resolveStep returns a copy of s with pipeline defaults merged in and every
// ${{ }} substituted in the string-valued fields the runner cares about.
func (sch *scheduler) resolveStep(s *Step, r *Resolver) (*Step, error) {
	out := *s // shallow copy; nested maps replaced below
	mergeDefaults(&out, sch.p.Defaults)

	if v, err := r.ResolveString(s.Format); err != nil {
		return nil, err
	} else {
		out.Format = v
	}
	if v, err := r.ResolveString(s.StatusCode); err != nil {
		return nil, err
	} else {
		out.StatusCode = v
	}
	if s.Params != nil {
		resolved, err := r.ResolveAny(anyMap(s.Params))
		if err != nil {
			return nil, err
		}
		out.Params = resolved.(map[string]any)
	}
	if s.Body != nil {
		resolved, err := r.ResolveAny(s.Body)
		if err != nil {
			return nil, err
		}
		out.Body = resolved
	}
	return &out, nil
}

// anyMap is a type helper to reuse ResolveAny on Step.Params.
func anyMap(m map[string]any) map[string]any { return m }

// mergeDefaults fills unset pointer fields on s from d. Step-level values
// always win when explicitly set (non-nil pointer / non-zero string).
func mergeDefaults(s *Step, d *StepDefaults) {
	if d == nil {
		return
	}
	if s.Verbose == nil {
		s.Verbose = d.Verbose
	}
	if s.IgnoreErrors == nil {
		s.IgnoreErrors = d.IgnoreErrors
	}
	if s.RetryCount == nil {
		s.RetryCount = d.RetryCount
	}
	if s.Parallelism == nil {
		s.Parallelism = d.Parallelism
	}
	if s.StatusCode == "" && d.StatusCode != nil {
		s.StatusCode = *d.StatusCode
	}
	if s.BasePath == nil {
		s.BasePath = d.BasePath
	}
	if s.Scheme == nil {
		s.Scheme = d.Scheme
	}
	if s.Server == nil {
		s.Server = d.Server
	}
	if s.Jwt == nil {
		s.Jwt = d.Jwt
	}
}

// extractOutputs runs each jq expression against the last record in the
// step's response. Multi-record steps use the last emitted record — see
// docs/pipeline.md for the recommendation to prefer inputFrom for bulk data.
func extractOutputs(records []Record, outputs map[string]string) (map[string]any, error) {
	if len(outputs) == 0 || len(records) == 0 {
		return nil, nil
	}
	last := records[len(records)-1]
	out := make(map[string]any, len(outputs))
	for name, expr := range outputs {
		q, err := gojq.Parse(expr)
		if err != nil {
			return nil, fmt.Errorf("output %q: parse %q: %w", name, expr, err)
		}
		iter := q.Run(any(last))
		v, ok := iter.Next()
		if !ok {
			continue
		}
		if e, isErr := v.(error); isErr {
			return nil, fmt.Errorf("output %q: %w", name, e)
		}
		out[name] = v
	}
	return out, nil
}

// truthy mirrors jq expression truthiness: nil / false /
// zero / empty string / empty collection is false; anything else is true.
func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != "" && x != "false" && x != "0"
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	default:
		return true
	}
}
