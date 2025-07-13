// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package parser

import (
	"fmt"
	"openapi/internal/common"
)

func (p *Parameter) DefaultStr() string {
	switch p.Default.(type) {
	case string:
		return fmt.Sprintf("%s", p.Default)
	case int:
		return fmt.Sprintf("%d", p.Default)
	case float64:
		return fmt.Sprintf("%f", p.Default)
	default:
		if p.Default == nil {
			return ""
		}
		return fmt.Sprintf("%s", p.Default)
	}
}

func (p *Parameter) GetDescription() string {
	d := p.Description
	if p.Schema != nil && p.Schema.Definition != nil {
		r := common.GetJsonString(p.Schema.Definition.Required)
		j := common.GetJsonString(p.Schema.Definition.Properties)
		d = fmt.Sprintf("%s%s\nRequired: %sProperties: %s",
			d, p.Schema.Ref, r, j)
	}
	return d
}

func printParameters(ps []Parameter) {
	for _, p := range ps {
		fmt.Println(p)
	}
}
