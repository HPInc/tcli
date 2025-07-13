// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package parser

const (
	InBody     = "body"
	InFormData = "formData"
	InPath     = "path"
	InQuery    = "query"
	InHeader   = "header"
)

// struct defs
type Schema struct {
	Ref        string      `json:"$ref"`
	Definition *Definition `json:"-"`
}

type Parameter struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	In          string      `json:"in"`
	Type        string      `json:"type,omitempty"`
	Format      string      `json:"format,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

type Extension struct {
	Class string `json:"class"`
}

type Method struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	OperationId string      `json:"operationId"`
	Tags        []string    `json:"tags"`
	Parameters  []Parameter `json:"parameters"`
	MethodName  string
	Consumes    []string `json:"consumes"`
	Path        string
	Extension   *Extension            `json:"extension,omitempty"`
	Securities  []map[string][]string `json:"security"`
}

type Path struct {
	Get     *Method `json:"get,omitempty"`
	Delete  *Method `json:"delete,omitempty"`
	Options *Method `json:"options,omitempty"`
	Patch   *Method `json:"patch,omitempty"`
	Post    *Method `json:"post,omitempty"`
	Put     *Method `json:"put,omitempty"`
}

type ArrayItem struct {
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
	Ref    string `json:"$ref,omitempty"`
}

type Property struct {
	Type        string     `json:"type,omitempty"`
	Format      string     `json:"format,omitempty"`
	Description string     `json:"description,omitempty"`
	Ref         string     `json:"$ref,omitempty"`
	Items       *ArrayItem `json:"items,omitempty"`
}

type Definition struct {
	Type       string               `json:"type"`
	Required   []string             `json:"required,omitempty"`
	Properties map[string]*Property `json:"properties"`
}

// top level commands / sub commands
// this could either be a tag or a direct command
type Command struct {
	Name        string
	Description string
	IsTag       bool
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SecurityDefinition struct {
	Type string `json:"type"`
}

type Root struct {
	Host                string                        `json:"host"`
	BasePath            string                        `json:"basePath"`
	Schemes             []string                      `json:"schemes"`
	Paths               map[string]*Path              `json:"paths,omitempty"`
	Tags                []Tag                         `json:"tags"`
	Definitions         map[string]Definition         `json:"definitions"`
	SecurityDefinitions map[string]SecurityDefinition `json:"securityDefinitions"`
}
