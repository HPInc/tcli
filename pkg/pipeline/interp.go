// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"regexp"
	"strings"
)

// interpolationExpr matches ${{ }} substitution sites in string values.
// Content is captured non-greedily so multiple expressions on one line resolve
// independently. Whitespace inside the braces is trimmed by the caller.
var interpolationExpr = regexp.MustCompile(`\$\{\{\s*([^}]+?)\s*\}\}`)

// Reference is one parsed ${{ ... }} occurrence in a pipeline value.
type Reference struct {
	Raw  string   // the literal "${{ ... }}" including braces, for replacement
	Path []string // dotted path split on "."
}

// FindReferences returns every ${{ }} occurrence inside s.
func FindReferences(s string) []Reference {
	matches := interpolationExpr.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return nil
	}
	refs := make([]Reference, 0, len(matches))
	for _, m := range matches {
		full := s[m[0]:m[1]]
		body := strings.TrimSpace(s[m[2]:m[3]])
		refs = append(refs, Reference{Raw: full, Path: strings.Split(body, ".")})
	}
	return refs
}

// Kind reports the top-level scope of a reference path: "variables", "steps",
// "env", or "" for anything else. Callers use this to route lookup.
func (r Reference) Kind() string {
	if len(r.Path) == 0 {
		return ""
	}
	switch r.Path[0] {
	case "variables", "steps", "env":
		return r.Path[0]
	}
	return ""
}

// StepDep returns the step name a reference depends on, or "" if the reference
// does not target a step. Used by the DAG builder to derive edges from
// ${{ steps.<name>.xx }} occurrences.
func (r Reference) StepDep() string {
	if r.Kind() == "steps" && len(r.Path) >= 2 {
		return r.Path[1]
	}
	return ""
}
