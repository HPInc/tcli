// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import "sync"

// StepStatus is the terminal (or in-progress) status of one step during a run.
type StepStatus int

const (
	StatusPending StepStatus = iota
	StatusRunning
	StatusSucceeded
	StatusFailed
	StatusSkipped
)

func (s StepStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusSucceeded:
		return "succeeded"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// StepResult captures everything about one completed step that later steps
// or the caller can observe.
type StepResult struct {
	Status  StepStatus
	Records []Record       // records the step emitted (may be nil if none)
	Outputs map[string]any // extracted per the step's outputs: block
	Err     error
}

// Ok reports whether the step reached StatusSucceeded. Used for
// ${{ steps.X.ok }} references and parent-status checks.
func (r *StepResult) Ok() bool { return r.Status == StatusSucceeded }

// State is the shared runtime state for one pipeline execution: variables
// (immutable snapshot from the YAML) and per-step results (mutated as steps
// complete). Reads and writes are safe from multiple goroutines.
type State struct {
	Variables map[string]any

	mu    sync.RWMutex
	steps map[string]*StepResult
}

// NewState seeds pending StepResults for every step so ${{ steps.X.xx }}
// references never hit a nil lookup.
func NewState(p *Pipeline) *State {
	s := &State{
		Variables: p.Variables,
		steps:     make(map[string]*StepResult, len(p.Steps)),
	}
	for _, step := range p.Steps {
		s.steps[step.Name] = &StepResult{Status: StatusPending}
	}
	return s
}

// Get returns the result for one step. Returns nil if the step name is unknown.
func (s *State) Get(name string) *StepResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.steps[name]
}

// Set replaces a step's result atomically.
func (s *State) Set(name string, r *StepResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.steps[name]; ok {
		s.steps[name] = r
	}
}

// Snapshot returns a shallow copy of the step results map. Useful for
// end-of-run reporting.
func (s *State) Snapshot() map[string]*StepResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]*StepResult, len(s.steps))
	for k, v := range s.steps {
		out[k] = v
	}
	return out
}
