// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"openapi/internal/common"
	"openapi/internal/parser"

	"gopkg.in/yaml.v2"
)

var currModule *Module

func GetCurrentModule() *Module {
	return currModule
}

func LoadModules(configFile string) bool {
	logger.Debug("loading module config file: ", configFile)

	// Open the configuration file for parsing.
	bytes, err := common.ReadFile(configFile)
	if err != nil {
		logger.Error("Failed to load module config file: ",
			configFile, ", Error = ", err)
		return false
	}

	var cfg ModuleConfig
	// Read the configuration file and unmarshal the YAML.
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		logger.Error("Failed to parse module config file: ",
			configFile, ", Error = ", err)
		return false
	}
	settings.Modules = cfg.Modules

	return true
}

func (m *Module) LoadConfig() error {
	var err error
	path := filepath.Join(getConfigRoot(), m.ConfigFile)
	m.ConfigRoot, err = parser.ReadSwagger(path)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func GetModule(module string) *Module {
	for _, m := range GetSettings().Modules {
		if m.Name == module {
			return &m
		}
	}
	return nil
}

func ShowModules() {
	maxNameLength := 0
	for _, m := range GetSettings().Modules {
		if len(m.Name) > maxNameLength {
			maxNameLength = len(m.Name)
		}
	}
	sepSpaces := maxNameLength + 4
	fmt.Println("Please specify a module. Supported modules are:")
	for _, m := range GetSettings().Modules {
		space := strings.Repeat(" ", sepSpaces-len(m.Name))
		logger.GetPlainLogger().Printf("\x1b[0;32m- %v%s%v \x1b[0m\n",
			m.Name, space, m.Description)
	}
}

// show supported commands in the specified module
func ShowCommands(module string) error {
	m := GetModule(module)
	if m == nil {
		logger.Errorf("Module \"%s\" not found\n", module)
	}
	if m != nil && m.LoadConfig() == nil {
		ShowModuleCommands(m)
		return nil
	}
	return common.ErrNoSuchModule
}

func ShowCommand(module, cmd, subCmd string) (*parser.Method, error) {
	currModule = GetModule(module)
	if currModule == nil {
		logger.Errorf("Module \"%s\" not found\n", module)
		return nil, common.ErrNoSuchModule
	} else if currModule.LoadConfig() == nil {
		for _, c := range currModule.ConfigRoot.GetCommands() {
			if c.Name == cmd {
				if c.IsTag {
					m, err := currModule.ConfigRoot.
						GetOperation(cmd, subCmd)
					if err != nil {
						return nil, common.ErrNoSuchSubCommand
					} else {
						return m, nil
					}
				} else {
					m, err := currModule.ConfigRoot.
						GetOperation("", cmd)
					if err != nil {
						return nil, common.ErrNoSuchCommand
					} else {
						return m, nil
					}
				}
			}
		}
	}
	return nil, common.ErrNoSuchCommand
}

func ShowModuleCommands(m *Module) {
	fmt.Println("Please specify a command. Supported commands are:")
	for _, c := range m.ConfigRoot.GetCommands() {
		fmt.Printf("\x1b[0;32m- %v\t%v \x1b[0m\n",
			c.Name, c.Description)
	}
}

func ShowTaggedCommands(cmd string) {
	fmt.Println("Please specify a sub command. Supported commands are:")
	cmds, err := currModule.ConfigRoot.GetTaggedCommands(cmd)
	if err == nil {
		for _, m := range cmds {
			fmt.Printf("\x1b[0;32m- %v\t%v \x1b[0m\n",
				m.OperationId, m.Summary)
		}
	}
}
