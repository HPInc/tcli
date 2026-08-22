// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/hpinc/tcli/pkg/common"
)

// Load reads and parses a pipeline YAML file, then runs full validation
// (schema, references, DAG cycles). A returned *Pipeline is safe to execute.
func Load(path string) (*Pipeline, error) {
	// common.ReadFile applies filepath.Clean to satisfy gosec G304 the
	// path is user-supplied by design (that's the API contract of Load).
	data, err := common.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read pipeline %q: %w", path, err)
	}
	return Parse(data)
}

// Parse decodes YAML bytes into a *Pipeline and validates it.
func Parse(data []byte) (*Pipeline, error) {
	var p Pipeline
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true) // reject unknown fields early feedback on typos
	if err := dec.Decode(&p); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("parse pipeline: %w", err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	// BuildDAG runs cycle detection; discarding the result is fine because
	// it's cheap to rebuild and callers usually want a fresh graph anyway.
	if _, err := BuildDAG(&p); err != nil {
		return nil, err
	}
	return &p, nil
}
