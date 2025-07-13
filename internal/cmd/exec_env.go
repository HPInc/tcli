// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"openapi/internal/common"
	"openapi/internal/config"
	"openapi/internal/parser"
)

type Environment struct {
	Args     []string
	Parallel *ParallelEnvironment
}

func GetExecutionEnv(args []string) *Environment {
	return &Environment{
		Args: args,
	}
}

func (e *Environment) Exec(m *parser.Method, i *common.Input) error {
	c, err := e.getCmd(m, i)
	if err != nil {
		return err
	}
	// this is here because we do not have early indication
	// on parallel exec env till we parse cmd line args
	if c.global.Parallel {
		if e.Parallel == nil {
			e.Parallel = newParallel()
		}
	}
	for i := uint(1); i <= c.global.Count; i++ {
		e.addJob(c)
	}
	return nil
}

func (e *Environment) Wait() error {
	if e.Parallel != nil {
		e.Parallel.wait()
	}
	return nil
}

func (e *Environment) addJob(c *CmdBase) {
	if e.Parallel != nil {
		e.Parallel.addJob(c)
	} else {
		_ = c.runFunc()
	}
}

func (e *Environment) getCmd(m *parser.Method, i *common.Input) (*CmdBase, error) {
	var err error
	p := New(m, i, config.GetCurrentModule().ConfigRoot)
	if err = p.Flags.Parse(e.Args); err == nil {
		if err = p.ValidateParams(); err != nil {
			return nil, err
		}
		// if there is an extension, it points to an implementation
		// such as sqs or mqtt to handle
		var c Command
		if c, err = GetCommand(m.GetExtensionClass()); err == nil {
			k := c.Init(p)
			return k.GetBase(), nil
		}
	}
	return nil, err
}
