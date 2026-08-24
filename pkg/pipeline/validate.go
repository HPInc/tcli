// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"errors"
	"fmt"
	"strings"
)

// Validate runs all structural checks on a pipeline. It returns the first
// error encountered; callers can inspect the wrapped error for detail.
// A pipeline that passes Validate is safe to build a DAG from.
func (p *Pipeline) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("pipeline: name is required")
	}
	if len(p.Steps) == 0 {
		return errors.New("pipeline: at least one step is required")
	}
	if p.Concurrency != nil {
		switch p.Concurrency.OnFailure {
		case "", OnFailureCancel, OnFailureDrain:
		default:
			return fmt.Errorf("pipeline: concurrency.onFailure must be %q or %q, got %q",
				OnFailureCancel, OnFailureDrain, p.Concurrency.OnFailure)
		}
		if p.Concurrency.MaxParallel < 0 {
			return fmt.Errorf("pipeline: concurrency.maxParallel must be >= 0, got %d",
				p.Concurrency.MaxParallel)
		}
	}

	names := make(map[string]struct{}, len(p.Steps))
	for _, s := range p.Steps {
		if err := validateStepShape(s); err != nil {
			return err
		}
		if _, dup := names[s.Name]; dup {
			return fmt.Errorf("step %q: duplicate name", s.Name)
		}
		names[s.Name] = struct{}{}
	}

	// Second pass: resolve step-to-step references now that all names are known.
	for _, s := range p.Steps {
		if s.InputFrom != "" {
			if _, ok := names[s.InputFrom]; !ok {
				return fmt.Errorf("step %q: inputFrom references unknown step %q",
					s.Name, s.InputFrom)
			}
			if s.InputFrom == s.Name {
				return fmt.Errorf("step %q: inputFrom references itself", s.Name)
			}
		}
		for _, dep := range s.DependsOn {
			if _, ok := names[dep]; !ok {
				return fmt.Errorf("step %q: dependsOn references unknown step %q",
					s.Name, dep)
			}
			if dep == s.Name {
				return fmt.Errorf("step %q: dependsOn references itself", s.Name)
			}
		}
		for _, ref := range collectStepRefs(s) {
			if _, ok := names[ref]; !ok {
				return fmt.Errorf("step %q: ${{ steps.%s.… }} references unknown step",
					s.Name, ref)
			}
			if ref == s.Name {
				return fmt.Errorf("step %q: ${{ steps.%s.… }} references itself",
					s.Name, ref)
			}
		}
	}
	return nil
}

func validateStepShape(s *Step) error {
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("step: name is required")
	}
	if strings.TrimSpace(s.Command) == "" {
		return fmt.Errorf("step %q: command is required", s.Name)
	}
	// The command form is "<module> [<submodule>] <operation>"; at least
	// module + operation so 2 tokens minimum, 3 maximum.
	tokens := strings.Fields(s.Command)
	if n := len(tokens); n < 2 || n > 3 {
		return fmt.Errorf("step %q: command must be \"<module> [<submodule>] <operation>\", got %q",
			s.Name, s.Command)
	}
	for key := range s.Params {
		if _, reserved := reservedParamKeys[key]; reserved {
			return fmt.Errorf("step %q: %q is a reserved pipeline field and must be set at the step top level, not under params",
				s.Name, key)
		}
	}
	if s.Parallelism != nil && *s.Parallelism < 0 {
		return fmt.Errorf("step %q: parallelism must be >= 0, got %d", s.Name, *s.Parallelism)
	}
	if s.Count < 0 {
		return fmt.Errorf("step %q: count must be >= 0, got %d", s.Name, s.Count)
	}
	return nil
}

// collectStepRefs returns every step name a step references via ${{ steps.X.… }}
// across params values, format, condition, statusCode, body (if string).
func collectStepRefs(s *Step) []string {
	seen := make(map[string]struct{})
	add := func(name string) {
		if name != "" {
			seen[name] = struct{}{}
		}
	}
	scan := func(v string) {
		for _, ref := range FindReferences(v) {
			add(ref.StepDep())
		}
	}
	scan(s.Format)
	scan(s.Condition)
	scan(s.StatusCode)
	if body, ok := s.Body.(string); ok {
		scan(body)
	}
	for _, v := range s.Params {
		if str, ok := v.(string); ok {
			scan(str)
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	return out
}
