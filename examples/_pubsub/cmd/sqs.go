// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"log"
	"openapi/internal/common"
)

func init() {
	cmds["sqs"] = &SqsCommand{}
}

type SqsCommand struct {
	CmdBase
	Data     string
	Url      string
	Endpoint string
}

func (c *SqsCommand) Init(p *ParseResult) Command {
	k := SqsCommand{}
	k.Endpoint = *p.Values["endpoint"]
	k.Data = *p.Values["data"]
	k.Url = *p.Values["queue_url"]
	k.baseInit(p, k.send)
	return &k
}

// Function to publish the event to SNS
func (c *SqsCommand) send() error {
	data := common.GetJsonString(c.Data)
	cli, err := common.NewSQSClient(c.Endpoint)
	if err != nil {
		log.Fatalf("Publish to sqs failed: %v\n", err)
	}
	err = cli.Send(c.Url, data)
	if err != nil {
		log.Fatalf("Publish to sqs failed: %v\n", err)
	}
	return nil
}
