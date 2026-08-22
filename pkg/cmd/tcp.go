// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"net"

	"github.com/hpinc/tcli/pkg/utils"
)

type TcpCommand struct {
	CmdBase
}

func init() {
	RegisterCommand("tcp", &TcpCommand{})
}

func (c *TcpCommand) Init(p *ParseResult) Command {
	k := TcpCommand{}
	k.InitBase(p, k.doTcpRetry)
	return &k
}

func (c *TcpCommand) doTcpRetry() error {
	addr := fmt.Sprintf("%s", c.global.Server)
	log.Debugf("Connecting to %s\n", addr)
	utils.RetryWait(c.global.RetryCount, connect(c.global.Scheme, addr))
	return nil
}

func connect(scheme, addr string) func() bool {
	return func() bool {
		conn, err := net.Dial(scheme, addr)
		if err != nil {
			return false
		}
		defer closeConnection(conn)
		return true
	}
}

func closeConnection(c net.Conn) {
	if c != nil {
		if err := c.Close(); err != nil {
			log.Error("Error closing tcp conn: ", err)
		}
	}
}
