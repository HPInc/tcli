// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

type CmdBase struct {
	p       *ParseResult
	global  *GlobalResult
	runFunc fnRun
	stop    <-chan struct{} // closed when another job has failed
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
	c.p = p
	c.global = p.Global
	c.runFunc = f
}

// Params returns the parsed parameter values for this command, so external
// Command implementations can read parameters without needing direct
// access to CmdBase's unexported fields.
func (c *CmdBase) Params() *ParseResult {
	return c.p
}

// Stop returns a channel that is closed once another job in the run has
// failed. Long running commands, such as ones that retry, should give up
// when it closes. It is nil (never closes) outside an Environment.
func (c *CmdBase) Stop() <-chan struct{} {
	return c.stop
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
