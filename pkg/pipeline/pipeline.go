// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// Package pipeline implements the tcli pipeline (YAML) execution engine.
// See docs/pipeline.md for the schema reference.
package pipeline

const (
	OnFailureCancel = "cancel"
	OnFailureDrain  = "drain"
)

type Pipeline struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Variables   map[string]any `yaml:"variables,omitempty"`
	Concurrency *Concurrency   `yaml:"concurrency,omitempty"`
	Defaults    *StepDefaults  `yaml:"defaults,omitempty"`
	Steps       []*Step        `yaml:"steps"`
}

type Concurrency struct {
	MaxParallel int    `yaml:"maxParallel,omitempty"`
	OnFailure   string `yaml:"onFailure,omitempty"`
}

// StepDefaults holds pipeline-wide defaults. Fields are pointers so an
// unset default can be distinguished from an explicit zero value (e.g.
// `retryCount: 0` is not the same as "not specified").
type StepDefaults struct {
	Verbose      *bool   `yaml:"verbose,omitempty"`
	IgnoreErrors *bool   `yaml:"ignoreErrors,omitempty"`
	RetryCount   *int    `yaml:"retryCount,omitempty"`
	Parallelism  *int    `yaml:"parallelism,omitempty"`
	StatusCode   *string `yaml:"statusCode,omitempty"`
	BasePath     *string `yaml:"basePath,omitempty"`
	Scheme       *string `yaml:"scheme,omitempty"`
	Server       *string `yaml:"server,omitempty"`
	Jwt          *string `yaml:"jwt,omitempty"`
	Doc          *string `yaml:"doc,omitempty"`
}

type Step struct {
	Name    string         `yaml:"name"`
	Command string         `yaml:"command"`
	Params  map[string]any `yaml:"params,omitempty"`
	Body    any            `yaml:"body,omitempty"`
	Format  string         `yaml:"format,omitempty"`

	InputFrom string            `yaml:"inputFrom,omitempty"`
	Outputs   map[string]string `yaml:"outputs,omitempty"`

	DependsOn       []string `yaml:"dependsOn,omitempty"`
	Condition       string   `yaml:"condition,omitempty"`
	ContinueOnError bool     `yaml:"continueOnError,omitempty"`

	Count        int     `yaml:"count,omitempty"`
	Parallelism  *int    `yaml:"parallelism,omitempty"`
	RetryCount   *int    `yaml:"retryCount,omitempty"`
	IgnoreErrors *bool   `yaml:"ignoreErrors,omitempty"`
	StatusCode   string  `yaml:"statusCode,omitempty"`
	Verbose      *bool   `yaml:"verbose,omitempty"`
	BasePath     *string `yaml:"basePath,omitempty"`
	Scheme       *string `yaml:"scheme,omitempty"`
	Server       *string `yaml:"server,omitempty"`
	Jwt          *string `yaml:"jwt,omitempty"`
}

// reservedParamKeys are pipeline-schema control fields that must not appear
// under a step's params: map. See "Reserved keys" in docs/pipeline.md.
var reservedParamKeys = map[string]struct{}{
	"format":       {},
	"count":        {},
	"parallelism":  {},
	"verbose":      {},
	"retryCount":   {},
	"ignoreErrors": {},
	"statusCode":   {},
	"basePath":     {},
	"scheme":       {},
	"server":       {},
	"jwt":          {},
	// snake_case forms of the same flags, in case a user types them
	// out of habit from the CLI.
	"parallel":      {}, // the CLI flag -parallel maps to parallelism > 1
	"retry_count":   {},
	"ignore_errors": {},
	"status_code":   {},
	"base_path":     {},
}
