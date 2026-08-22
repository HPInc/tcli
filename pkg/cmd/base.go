// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

type CmdBase struct {
	p       *ParseResult
	global  *GlobalResult
	runFunc fnRun
}

func (c *CmdBase) baseInit(p *ParseResult, f fnRun) {
	c.p = p
	c.global = p.Global
	c.runFunc = f
}

// InitBase initializes a CmdBase from outside this package. Custom Command
// implementations registered via RegisterCommand should call this from
// their Init method instead of reimplementing CmdBase's plumbing:
//
//	func (c *MyCommand) Init(p *cmd.ParseResult) cmd.Command {
//	    k := MyCommand{}
//	    k.InitBase(p, k.run)
//	    return &k
//	}
func (c *CmdBase) InitBase(p *ParseResult, f func() error) {
	c.baseInit(p, f)
}

// Params returns the parsed parameter values for this command, so external
// Command implementations can read parameters without needing direct
// access to CmdBase's unexported fields.
func (c *CmdBase) Params() *ParseResult {
	return c.p
}

// Global returns the parsed global flag values for this command.
func (c *CmdBase) Global() *GlobalResult {
	return c.global
}

func (c *CmdBase) GetBase() *CmdBase {
	return c
}

func (c *CmdBase) Execute() error {
	return nil
}
