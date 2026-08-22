// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package env

import (
	"flag"
	"fmt"
	"os"

	"github.com/hpinc/tcli/pkg/pipeline"
)

// runPipeline dispatches `tcli pipeline <sub> ...` invocations. Kept in the
// env package so main.go stays a single entry point regardless of whether
// the user runs a module command or a pipeline.
func runPipeline(args []string) error {
	if len(args) == 0 {
		return showPipelineHelp()
	}
	switch args[0] {
	case "run":
		return runPipelineRun(args[1:])
	case "-h", "--help", "help":
		return showPipelineHelp()
	default:
		fmt.Fprintf(os.Stderr, "unknown pipeline subcommand %q\n\n", args[0])
		return showPipelineHelp()
	}
}

// runPipelineRun executes `tcli pipeline run [flags] <file.yaml>`.
func runPipelineRun(args []string) error {
	fs := flag.NewFlagSet("pipeline run", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: tcli pipeline run <file.yaml>")
	}
	file := fs.Arg(0)

	_, err := pipeline.Load(file)
	if err != nil {
		return err
	}
	return err
}

func showPipelineHelp() error {
	fmt.Println("Usage: tcli pipeline <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  run <file.yaml>   Execute a pipeline file")
	fmt.Println("  help              Show this help")
	return nil
}
