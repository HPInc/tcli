// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pubsub

import (
	"fmt"

	pubsubcommon "github.com/hpinc/tcli/examples/_pubsub/common"
	"github.com/hpinc/tcli/pkg/cmd"
)

func init() {
	cmd.RegisterCommand("sns", &SnsCommand{})
}

type SnsCommand struct {
	cmd.CmdBase
	Name     string
	Data     string
	Arn      string
	Endpoint string
}

func (c *SnsCommand) Init(p *cmd.ParseResult) cmd.Command {
	k := SnsCommand{}
	k.Endpoint = *p.Values["endpoint"]
	k.Data = *p.Values["data"]
	k.Arn = *p.Values["arn"]
	k.Name = *p.Values["name"]
	k.InitBase(p, k.send)
	return &k
}

// Function to publish the event to SNS
func (c *SnsCommand) send() error {
	cli, err := pubsubcommon.NewSNSClient(c.Endpoint)
	if err != nil {
		return fmt.Errorf("publish to sns failed: %w", err)
	}
	if err := cli.Publish(c.Arn, c.Data, c.Name); err != nil {
		return fmt.Errorf("publish to sns failed: %w", err)
	}
	return nil
}
