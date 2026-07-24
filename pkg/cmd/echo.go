// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

type EchoCommand struct {
	CmdBase
}

func init() {
	RegisterCommand("echo", &EchoCommand{})
}

func (c *EchoCommand) Init(p *ParseResult) Command {
	k := EchoCommand{}
	k.InitBase(p, k.doEcho)
	return &k
}

func (c *EchoCommand) doEcho() error {
	applyFormat([]byte(*c.p.Values["data"]), c.p)
	return nil
}
