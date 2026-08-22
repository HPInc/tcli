// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package env

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/hpinc/tcli/pkg/cmd"
	"github.com/hpinc/tcli/pkg/common"
	"github.com/hpinc/tcli/pkg/config"
	"github.com/hpinc/tcli/pkg/parser"
)

const (
	cmdModules  = 1
	cmdCommands = 2
	cmdCommand  = 3

	//
	posModule     = 1
	posCommand    = 2
	posSubCommand = 3
)

type cmdState struct {
	state  int
	module string
	cmd    string
	subCmd string
	args   []string
}

type module struct {
	Tag parser.Tag
}

var (
	modules []module
	argc    = len(os.Args)
	state   = getCmdState()
)

// Run is the main entry point for executing commands based on command line arguments
func Run() error {
	var err error
	if !config.Load() {
		return errors.New("config load failed")
	}
	// handle command line args
	switch state.state {
	case cmdModules:
		config.ShowModules()
	case cmdCommands:
		err = config.ShowCommands(state.module)
	case cmdCommand:
		err = call(config.ShowCommand(state.module, state.cmd, state.subCmd))
	}
	return handleError(err)
}

// call executes the given method with the provided error handling
func call(m *parser.Method, err error) error {
	if err != nil {
		return err
	}

	env := cmd.GetExecutionEnv(state.args)

	// read from stdin and feed to commands
	if hasStdin() {
		input := json.NewDecoder(os.Stdin)
		var any common.Input
		for err == nil {
			err = input.Decode(&any)
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				break
			}
			err = env.Exec(m, &any)
		}
	} else {
		err = env.Exec(m, nil)
	}
	return env.Wait()
}

// getCmdState determines the current command state based on command line arguments
func getCmdState() cmdState {
	if argc < 2 {
		return cmdState{state: cmdModules}
	} else if argc < 3 {
		return cmdState{state: cmdCommands, module: os.Args[posModule]}
	} else {
		subCmd := getSubCmd()
		args := os.Args[3:]
		if subCmd != "" {
			args = os.Args[4:]
		}
		return cmdState{
			state:  cmdCommand,
			module: os.Args[posModule],
			cmd:    os.Args[posCommand],
			subCmd: subCmd,
			args:   args}
	}
}

// subcommands are present when an api has a tagged set of commands
// this will be the final level of supported drill in for commands
func getSubCmd() string {
	if argc > posSubCommand {
		temp := os.Args[posSubCommand]
		if !strings.HasPrefix(temp, "-") {
			return temp
		}
	}
	return ""
}

// Attempt to avoid specifying a flag for stdin
// check if there is something readable at stdin
// this is not tested beyond minimal redirection
// it is tested in windows as well but there could be
// corner cases when it doesnt work. If that is the case
// it's better to specify a flag to indicate input
func hasStdin() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

// handleError processes errors and displays appropriate information
func handleError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, common.ErrNoSuchModule) {
		config.ShowModules()
	} else if errors.Is(err, common.ErrNoSuchCommand) {
		err = config.ShowCommands(state.module)
	} else if errors.Is(err, common.ErrNoSuchSubCommand) {
		config.ShowTaggedCommands(state.cmd)
	} else {
		return err
	}
	return nil
}
