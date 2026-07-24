// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// Package pubsub demonstrates adding a custom extension command to tcli
// without vendoring or copying files into the tcli module itself. It
// imports tcli's public pkg/cmd package and self-registers via
// cmd.RegisterCommand. Consumers only need to blank-import this package
// (e.g. `import _ "example.com/myclient/pubsub"`) from their main so its
// init() runs.
package pubsub

import (
	"fmt"

	pubsubcommon "github.com/hpinc/tcli/examples/_pubsub/common"
	"github.com/hpinc/tcli/pkg/cmd"
	tcommon "github.com/hpinc/tcli/pkg/common"
)

func init() {
	cmd.RegisterCommand("sqs", &SqsCommand{})
}

type SqsCommand struct {
	cmd.CmdBase
	Data     string
	Url      string
	Endpoint string
}

func (c *SqsCommand) Init(p *cmd.ParseResult) cmd.Command {
	k := SqsCommand{}
	k.Endpoint = *p.Values["endpoint"]
	k.Data = *p.Values["data"]
	k.Url = *p.Values["queue_url"]
	k.InitBase(p, k.send)
	return &k
}

// Function to publish the event to SQS
func (c *SqsCommand) send() error {
	data := tcommon.GetJsonString(c.Data)
	cli, err := pubsubcommon.NewSQSClient(c.Endpoint)
	if err != nil {
		return fmt.Errorf("publish to sqs failed: %w", err)
	}
	if err := cli.Send(c.Url, data); err != nil {
		return fmt.Errorf("publish to sqs failed: %w", err)
	}
	return nil
}
