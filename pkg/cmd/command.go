// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/hpinc/tcli/pkg/common"
	"github.com/hpinc/tcli/pkg/config"
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

// RegisterCommand registers a Command implementation under the given
// extension class name (the value returned by Method.GetExtensionClass(),
// which maps to the "class" field of an operation's extension in the
// OpenAPI spec).
//
// This lets a program that imports tcli as a library add its own command
// types (e.g. for publishing to a queue or topic) without needing to copy
// source files into this package. Call it from an init() in your own
// package, then blank-import that package from your main so the init()
// runs:
//
//	package mypubsub
//
//	func init() {
//	    cmd.RegisterCommand("sqs", &SqsCommand{})
//	}
//
//	// in your main package
//	import _ "example.com/myclient/mypubsub"
func RegisterCommand(name string, c Command) {
	cmds[name] = c
}
