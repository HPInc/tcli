// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"fmt"
	"os"
	"strings"
)

// Resolver evaluates ${{ }} references against a State. Runtime environment
// lookups go through Env; tests can substitute a fake to keep them hermetic.
type Resolver struct {
	State *State
	Env   func(string) string // defaults to os.Getenv when nil
}

// NewResolver returns a Resolver bound to state, using os.Getenv for
// ${{ env.X }} lookups.
func NewResolver(state *State) *Resolver {
	return &Resolver{State: state, Env: os.Getenv}
}

// ResolveString replaces every ${{ }} occurrence in s with its string
// value. Missing references return an error rather than silently substituting
// empty; missing values in configuration are almost always bugs.
func (r *Resolver) ResolveString(s string) (string, error) {
	for _, ref := range FindReferences(s) {
		val, err := r.resolve(ref)
		if err != nil {
			return "", err
		}
		s = strings.Replace(s, ref.Raw, fmt.Sprint(val), 1)
	}
	return s, nil
}

// ResolveValue returns the typed value when s is exactly one ${{ }}
// reference (so bools stay bools, numbers stay numbers), and falls back to
// string interpolation otherwise. Used where the resulting type matters —
// notably `condition:`.
func (r *Resolver) ResolveValue(s string) (any, error) {
	trimmed := strings.TrimSpace(s)
	refs := FindReferences(trimmed)
	if len(refs) == 1 && refs[0].Raw == trimmed {
		return r.resolve(refs[0])
	}
	return r.ResolveString(s)
}

// ResolveAny walks a value: for strings, interpolate; for maps/slices,
// recurse. Non-string leaves pass through unchanged.
func (r *Resolver) ResolveAny(v any) (any, error) {
	switch x := v.(type) {
	case string:
		return r.ResolveValue(x)
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			resolved, err := r.ResolveAny(vv)
			if err != nil {
				return nil, err
			}
			out[k] = resolved
		}
		return out, nil
	case []any:
		out := make([]any, len(x))
		for i, vv := range x {
			resolved, err := r.ResolveAny(vv)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		}
		return out, nil
	default:
		return v, nil
	}
}

func (r *Resolver) resolve(ref Reference) (any, error) {
	if len(ref.Path) == 0 {
		return nil, fmt.Errorf("empty reference %q", ref.Raw)
	}
	switch ref.Path[0] {
	case "variables":
		if len(ref.Path) < 2 {
			return nil, fmt.Errorf("variables reference needs a name: %q", ref.Raw)
		}
		val, ok := r.State.Variables[ref.Path[1]]
		if !ok {
			return nil, fmt.Errorf("unknown variable %q", ref.Path[1])
		}
		return walkPath(val, ref.Path[2:])
	case "env":
		if len(ref.Path) < 2 {
			return nil, fmt.Errorf("env reference needs a name: %q", ref.Raw)
		}
		getenv := r.Env
		if getenv == nil {
			getenv = os.Getenv
		}
		return getenv(ref.Path[1]), nil
	case "steps":
		return r.resolveStep(ref)
	default:
		return nil, fmt.Errorf("unknown reference scope %q in %q", ref.Path[0], ref.Raw)
	}
}

func (r *Resolver) resolveStep(ref Reference) (any, error) {
	if len(ref.Path) < 3 {
		return nil, fmt.Errorf("steps reference needs step and field: %q", ref.Raw)
	}
	stepName, field := ref.Path[1], ref.Path[2]
	step := r.State.Get(stepName)
	if step == nil {
		return nil, fmt.Errorf("unknown step %q", stepName)
	}
	switch field {
	case "ok":
		return step.Ok(), nil
	case "outputs":
		if len(ref.Path) < 4 {
			return nil, fmt.Errorf("outputs reference needs a name: %q", ref.Raw)
		}
		val, ok := step.Outputs[ref.Path[3]]
		if !ok {
			return nil, fmt.Errorf("step %q has no output named %q", stepName, ref.Path[3])
		}
		return walkPath(val, ref.Path[4:])
	default:
		// ${{ steps.X.status }} was in the schema draft but the subprocess
		// runner cannot report HTTP status codes deferred to v2.
		return nil, fmt.Errorf("unsupported step field %q in %q", field, ref.Raw)
	}
}

// walkPath descends into a nested map by dotted field name. Used to support
// references like ${{ variables.config.timeout }} against nested YAML values.
func walkPath(v any, path []string) (any, error) {
	for _, p := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot descend into %q: not an object", p)
		}
		v = m[p]
	}
	return v, nil
}
