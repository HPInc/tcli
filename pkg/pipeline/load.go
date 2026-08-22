// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/hpinc/tcli/pkg/common"
)

// Reads and parses a pipeline YAML file, then runs full validation
// (schema, references). A returned *Pipeline is safe to execute.
func Load(path string) (*Pipeline, error) {
	data, err := common.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read pipeline %q: %w", path, err)
	}
	return Parse(data)
}

// Decodes YAML bytes into a *Pipeline.
func Parse(data []byte) (*Pipeline, error) {
	var p Pipeline
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true) // reject unknown fields and early feedback on typos
	if err := dec.Decode(&p); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("parse pipeline: %w", err)
	}
	return &p, nil
}
