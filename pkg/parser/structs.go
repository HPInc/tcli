// Copyright (c) 2025-2026 HP Development Company, L.P.
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
	Ref        string
	Definition *Definition
}

type Parameter struct {
	Name        string
	Description string
	Required    bool
	In          string
	Type        string
	Format      string
	Schema      *Schema
	Default     interface{}
}

type Extension struct {
	Class string
}

type Method struct {
	Summary     string
	Description string
	OperationId string
	Tags        []string
	Parameters  []Parameter
	MethodName  string
	Consumes    []string
	Path        string
	Extension   *Extension
	Securities  []map[string][]string
}

type Path struct {
	Get     *Method
	Delete  *Method
	Options *Method
	Patch   *Method
	Post    *Method
	Put     *Method
}

type ArrayItem struct {
	Type   string
	Format string
	Ref    string
}

type Property struct {
	Type        string
	Format      string
	Description string
	Ref         string
	Items       *ArrayItem
}

type Definition struct {
	Type       string
	Required   []string
	Properties map[string]*Property
}

// top level commands / sub commands
// this could either be a tag or a direct command
type Command struct {
	Name        string
	Description string
	IsTag       bool
}

type Tag struct {
	Name        string
	Description string
}

type SecurityDefinition struct {
	Type string
}

type Root struct {
	Host                string
	BasePath            string
	Schemes             []string
	Paths               map[string]*Path
	Tags                []Tag
	Definitions         map[string]Definition
	SecurityDefinitions map[string]SecurityDefinition
}
