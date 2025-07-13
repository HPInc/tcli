// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import "openapi/internal/parser"

// proxy config
type Proxy struct {
	HttpProxy  string `yaml:"http_proxy"`
	HttpsProxy string `yaml:"https_proxy"`
}

// server addresses for destination servers
type Server struct {
	Addresses map[string]string `yaml:"addresses"`
}

// profile bag holding all configs
type Profile struct {
	Name       string `yaml:"name"`
	ModulesDir string `yaml:"modules_dir"`
}

// module config
type Module struct {
	Name        string `yaml:"name"`
	ConfigFile  string `yaml:"config"`
	Description string `yaml:"description"`
	ConfigRoot  *parser.Root
}

// Config holds all profiles and overrides to current profile
type Config struct {
	ProfileName    string    `yaml:"default_profile"`
	Profiles       []Profile `yaml:"profiles"`
	ModuleConfig   string    `yaml:"module_config"`
	CurrentProfile *Profile
	Modules        []Module
	TokenCache     *TokenCache
}

type ModuleConfig struct {
	Modules []Module `yaml:"modules"`
}
