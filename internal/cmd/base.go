// Copyright 2025 HP Development Company, L.P.
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

func (c *CmdBase) GetBase() *CmdBase {
	return c
}

func (c *CmdBase) Execute() error {
	return nil
}
