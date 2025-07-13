// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"log"
	"openapi/internal/common"
)

func init() {
	cmds["sns"] = &SnsCommand{}
}

type SnsCommand struct {
	CmdBase
	Name     string
	Data     string
	Arn      string
	Endpoint string
}

func (c *SnsCommand) Init(p *ParseResult) Command {
	k := SnsCommand{}
	k.Endpoint = *p.Values["endpoint"]
	k.Data = *p.Values["data"]
	k.Arn = *p.Values["arn"]
	k.Name = *p.Values["name"]
	k.baseInit(p, k.send)
	return &k
}

// Function to publish the event to SNS
func (c *SnsCommand) send() error {
	cli, err := common.NewSNSClient(c.Endpoint)
	if err != nil {
		log.Fatalf("Publish to sns failed: %v\n", err)
	}
	err = cli.Publish(c.Arn, c.Data, c.Name)
	if err != nil {
		log.Fatalf("Publish to sns failed: %v\n", err)
	}
	return nil
}
