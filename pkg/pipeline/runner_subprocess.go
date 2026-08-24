// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// SubprocessRunner executes each step by invoking the tcli binary as a
// subprocess. This deliberately reuses the CLI code path same
// argument-parsing, same module/command dispatch, same stdout format so
// scenario behavior is identical to what the equivalent bash pipe would do
// today. In-process execution is deferred to a later iteration.
//
// The pipeline's `parallelism` field is translated best-effort: for any
// value other than unset/1 we pass `-parallel` (tcli's boolean flag). Exact
// worker-count control requires an eventual `-parallelism <n>` addition
// to pkg/cmd.
type SubprocessRunner struct {
	// Binary is the path to the tcli executable. When empty, os.Executable()
	// is used, which is correct when the current process itself is tcli.
	Binary string

	// Stderr, if set, is where the child's stderr is copied. Defaults to
	// os.Stderr so tcli's logs surface during a pipeline run.
	Stderr io.Writer
}

// Run invokes tcli for one step, streams input records on stdin, captures
// stdout, and parses the JSON stream back into records.
func (r *SubprocessRunner) Run(ctx context.Context, s *Step, input []Record) ([]Record, error) {
	binary := r.Binary
	if binary == "" {
		self, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("locate tcli binary: %w", err)
		}
		binary = self
	}

	args, err := BuildArgs(s)
	if err != nil {
		return nil, err
	}

	// #nosec G204 -- shelling out to tcli is the runner's purpose; binary is
	// resolved via os.Executable() (or an explicitly-set field), and args are
	// derived from a Step that has already passed pipeline schema validation.
	cmd := exec.CommandContext(ctx, binary, args...)
	stderr := r.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	cmd.Stderr = stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if len(input) > 0 {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, fmt.Errorf("stdin pipe: %w", err)
		}
		go func() {
			defer func() { _ = stdin.Close() }()
			enc := json.NewEncoder(stdin)
			for _, rec := range input {
				_ = enc.Encode(rec)
			}
		}()
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tcli %s: %w", strings.Join(args, " "), err)
	}
	return parseRecords(stdout.Bytes())
}

// BuildArgs turns a resolved Step into the CLI argument vector tcli expects.
// Exposed so callers writing custom runners can reuse the mapping.
func BuildArgs(s *Step) ([]string, error) {
	tokens := strings.Fields(s.Command)
	if n := len(tokens); n < 2 || n > 3 {
		return nil, fmt.Errorf("step %q: command must be \"<module> [<submodule>] <operation>\"", s.Name)
	}
	args := append([]string{}, tokens...)

	if s.Format != "" {
		args = append(args, "-format", s.Format)
	}
	if s.StatusCode != "" {
		args = append(args, "-status_code", s.StatusCode)
	}
	if s.Count > 0 {
		args = append(args, "-count", strconv.Itoa(s.Count))
	}
	if s.RetryCount != nil {
		args = append(args, "-retry_count", strconv.Itoa(*s.RetryCount))
	}
	if s.IgnoreErrors != nil && *s.IgnoreErrors {
		args = append(args, "-ignore_errors")
	}
	if s.Verbose != nil && *s.Verbose {
		args = append(args, "-v")
	}
	// parallelism 0 (auto) or >1 both map to tcli's boolean -parallel;
	// exact worker count is not currently propagatable.
	if s.Parallelism != nil && (*s.Parallelism == 0 || *s.Parallelism > 1) {
		args = append(args, "-parallel")
	}
	if s.BasePath != nil && *s.BasePath != "" {
		args = append(args, "-base_path", *s.BasePath)
	}
	if s.Scheme != nil && *s.Scheme != "" {
		args = append(args, "-scheme", *s.Scheme)
	}
	if s.Server != nil && *s.Server != "" {
		args = append(args, "-server", *s.Server)
	}
	if s.Jwt != nil && *s.Jwt != "" {
		args = append(args, "-jwt", *s.Jwt)
	}
	if s.Body != nil {
		body, err := encodeBody(s.Body)
		if err != nil {
			return nil, err
		}
		args = append(args, "-body", body)
	}
	for k, v := range s.Params {
		args = append(args, "-"+k, fmt.Sprint(v))
	}
	return args, nil
}

// encodeBody serializes step.body: strings pass through, everything else is
// JSON-encoded so tcli receives a well-formed body string.
func encodeBody(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("body: %w", err)
	}
	return string(b), nil
}

// parseRecords reads a stream of JSON values from tcli's stdout. Values
// come from applyFormat / fmt.Println, one per line. Objects become records
// directly; arrays are flattened so a list endpoint like findPetsByStatus
// yields one record per element (matching how a downstream jq expression
// would iterate them). Scalar top-level values are ignored — no established
// convention for turning a bare number/string into a record.
func parseRecords(data []byte) ([]Record, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	var out []Record
	for {
		var v any
		if err := dec.Decode(&v); err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, fmt.Errorf("decode tcli output: %w", err)
		}
		switch x := v.(type) {
		case map[string]any:
			out = append(out, x)
		case []any:
			for _, elem := range x {
				if m, ok := elem.(map[string]any); ok {
					out = append(out, m)
				}
			}
		}
	}
}
