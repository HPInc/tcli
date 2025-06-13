// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"openapi/internal/common"
	"openapi/internal/config"
)

// hold commands in a module
type commands map[string]Command
type fnRun func() error

type Command interface {
	Init(*ParseResult) Command
	Execute() error
	GetBase() *CmdBase
}

var (
	cmds = make(commands)
	log  = config.GetLogger()
)

func GetCommand(name string) (Command, error) {
	if c, ok := cmds[name]; ok {
		return c, nil
	}
	return nil, common.ErrNoSuchCommand
}
