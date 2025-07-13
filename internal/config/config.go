// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"openapi/internal/common"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	ConfigDir      = ".tcli"
	DataDir        = "data"
	EnvConfigRoot  = "TCLI_CONFIG_ROOT"
	EnvConfigFile  = "TCLI_CONFIG_FILE"
	ConfigFileName = "config.yaml"
	ModuleFileName = "modules.yaml"
)

var (
	DefaultHomeDir    = getHomeDir()
	DefaultConfigRoot = filepath.Join(DefaultHomeDir, ConfigDir)
	DefaultConfigFile = filepath.Join(getConfigRoot(), ConfigFileName)
	settings          Config
	logger            *Log
)

func init() {
	logger = InitLogger(getLogLevel())
}

// load config
func Load() bool {
	configFile := getConfigFile()
	logger.Debug("loading config file: ", configFile)

	// Open the configuration file for parsing.
	bytes, err := common.ReadFile(configFile)
	if err != nil {
		logger.Error("Failed to load configuration file: ",
			configFile, ", Error = ", err)
		return false
	}

	// Read the configuration file and unmarshal the YAML.
	err = yaml.Unmarshal(bytes, &settings)
	if err != nil {
		logger.Error("Failed to parse configuration file: ",
			configFile, ", Error = ", err)
		return false
	}

	settings.applyProfile()

	if LoadModules(settings.CurrentProfile.getModuleConfig()) {
		settings.TokenCache = NewTokenCache()
	} else {
		return false
	}
	return true
}

// read config file from cmdline args
// note: this will only work as an arg to the cli program
// not to the modules
func getConfigFile() string {
	configFile := os.Getenv(EnvConfigFile)
	if configFile == "" {
		configFile = DefaultConfigFile
	}
	return configFile
}

// provide a log util
func GetLogger() *Log {
	return logger
}

// return settings loaded from config
func GetSettings() *Config {
	return &settings
}

// set log level to verbose
func SetVerboseLogging() {
	logger.SetLevel(Verbose)
}

// set doc
func SetDocType(doc string) {
	logger.SetDocType(doc)
}

// return config root
// if TCLI_CONFIG_ROOT is set, root is the value of env
// if not set, root is user's home directory
// if there is a failure in loading user's home directory,
// current working directory is set as config root
func getConfigRoot() string {
	configRoot := os.Getenv(EnvConfigRoot)
	if configRoot == "" {
		configRoot = DefaultConfigRoot
	}
	return configRoot
}

// return current user home dir or pwd if that fails
func getHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		logger.Println("Failed to load user home dir: ", err)
		dir = "."
	}
	return dir
}

// return config dir
func getConfigDir() string {
	return filepath.Join(getConfigRoot())
}

// log level env
func getLogLevel() Level {
	var l Level = Info
	return l
}

// access token
func GetTokenCache() *TokenCache {
	return settings.TokenCache
}
