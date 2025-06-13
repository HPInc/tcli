// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"openapi/internal/common"
)

const definitionsPrefix = "#/definitions/"
const DefaultOperationClass = "http"
const OAuth2 = "oauth2"

func (r *Root) GetOperation(tag, operation string) (*Method, error) {
	for k, p := range r.Paths {
		for _, i := range []*Method{p.Get, p.Post, p.Patch, p.Delete, p.Put} {
			if i.hasTaggedOperation(tag, operation) {
				r.resolveDefinitions(i)
				if i == p.Get {
					i.MethodName = "GET"
				} else if i == p.Post {
					i.MethodName = "POST"
				} else if i == p.Patch {
					i.MethodName = "PATCH"
				} else if i == p.Delete {
					i.MethodName = "DELETE"
				} else if i == p.Put {
					i.MethodName = "PUT"
				}
				i.Path = k
				return i, nil
			}
		}
	}
	return nil, common.ErrNotFound
}

func (r *Root) resolveDefinitions(m *Method) {
	for _, p := range m.Parameters {
		if p.Schema != nil {
			s, ok := strings.CutPrefix(p.Schema.Ref, definitionsPrefix)
			if ok {
				if d, ok := r.Definitions[s]; ok {
					p.Schema.Definition = &d
				}
			}
		}
	}
}

func (r *Root) GetCommands() []*Command {
	var cmds []*Command
	// in case there are no root level tags
	// try to collect from paths
	if len(r.Tags) == 0 {
		r.Tags = r.collectTagsFromPaths()
	}

	for _, t := range r.Tags {
		cmds = append(cmds, &Command{
			Name:        t.Name,
			Description: t.Description,
			IsTag:       true})
	}
	for _, p := range r.Paths {
		for _, i := range []*Method{p.Get, p.Post, p.Patch, p.Delete, p.Put} {
			if i != nil && len(i.Tags) == 0 {
				cmds = append(cmds, &Command{
					Name:        i.OperationId,
					Description: fmt.Sprintf("%s\n%s", i.Summary, i.Description),
				})
			}
		}
	}
	return cmds
}

// get all paths with tags, return as Tags array
func (r *Root) collectTagsFromPaths() []Tag {
	tagsMap := make(map[string]bool)
	for _, p := range r.Paths {
		for _, i := range []*Method{p.Get, p.Post, p.Patch, p.Delete, p.Put} {
			if i != nil && len(i.Tags) != 0 {
				for _, t := range i.Tags {
					if _, ok := tagsMap[t]; !ok {
						tagsMap[t] = true
					}
				}
			}
		}
	}
	var tags []Tag
	for k, _ := range tagsMap {
		tags = append(tags, Tag{Name: k})
	}
	return tags
}

func (r *Root) GetTaggedCommands(tag string) ([]*Method, error) {
	var m []*Method
	for _, p := range r.Paths {
		for _, i := range []*Method{p.Get, p.Post, p.Patch, p.Delete, p.Put} {
			if i.hasOperation(tag) {
				m = append(m, i)
			}
		}
	}
	var err error
	if len(m) == 0 {
		err = common.ErrNotFound
	}
	return m, err
}

func (r *Root) ShowTags() {
	for _, i := range r.Tags {
		fmt.Println(">> Tag: ", i.Name, i.Description)
	}
}

func (r *Root) NeedJwt(s string) bool {
	for k, v := range r.SecurityDefinitions {
		if k == s && v.Type == OAuth2 {
			return true
		}
	}
	return false
}

func ReadSwagger(f string) (*Root, error) {
	bytes, err := os.ReadFile(f) // #nosec G304
	if err != nil {
		return nil, err
	}
	var r Root
	if err := json.Unmarshal(bytes, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (m *Method) hasOperation(tag string) bool {
	if m == nil {
		return false
	}
	for _, t := range m.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (m *Method) hasTaggedOperation(tag, operation string) bool {
	if m == nil {
		return false
	}
	// slight twist in meaning but simplifies caller
	// handles a lazy input def which specifies operations with no tag
	if tag == "" {
		return m.OperationId == operation
	}
	// this is the normal, we have a non-empty tag
	for _, t := range m.Tags {
		if t == tag && m.OperationId == operation {
			return true
		}
	}
	return false
}

func (m *Method) HasExtension() bool {
	if m == nil {
		return false
	}
	return m.Extension != nil
}

func (m *Method) GetExtensionClass() string {
	if m == nil || m.Extension == nil {
		return DefaultOperationClass
	}
	return m.Extension.Class
}

func (m *Method) HasSecurity() bool {
	return len(m.Securities) > 0
}

func (m *Method) NeedJwt(r *Root) bool {
	if m.HasSecurity() {
		for _, e := range m.Securities {
			for k, _ := range e {
				if r.NeedJwt(k) {
					return true
				}
			}
		}
	}
	return false
}

func printProperties(p map[string]*Property) {
	for k, v := range p {
		fmt.Println(k, v.Type, v.Format, v.Ref)
		if v.Type == "array" {
			fmt.Println(v.Items)
		}
	}
}
