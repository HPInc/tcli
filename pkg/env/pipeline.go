// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package env

import (
	"context"
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
	case "validate":
		return runPipelineValidate(args[1:])
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
	verbose := fs.Bool("v", false, "verbose per-step status")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: tcli pipeline run [-v] <file.yaml>")
	}
	file := fs.Arg(0)

	p, err := pipeline.Load(file)
	if err != nil {
		return err
	}
	executor := pipeline.NewExecutor(&pipeline.SubprocessRunner{})
	state, runErr := executor.Run(context.Background(), p)
	printPipelineResults(p, state, *verbose)
	return runErr
}

// runPipelineValidate parses and validates a pipeline file without running it.
// Useful for CI: catch schema / DAG errors before deployment.
func runPipelineValidate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tcli pipeline validate <file.yaml>")
	}
	p, err := pipeline.Load(args[0])
	if err != nil {
		return err
	}
	fmt.Printf("ok: %s (%d steps)\n", p.Name, len(p.Steps))
	return nil
}

func showPipelineHelp() error {
	fmt.Println("Usage: tcli pipeline <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  run [-v] <file.yaml>   Execute a pipeline file")
	fmt.Println("  validate <file.yaml>   Parse and validate a pipeline file")
	fmt.Println("  help                   Show this help")
	return nil
}

// printPipelineResults emits a one-line-per-step summary to stderr so it is
// visually separated from any JSON records that a step wrote to stdout.
func printPipelineResults(p *pipeline.Pipeline, state *pipeline.State, verbose bool) {
	if state == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "\npipeline %q:\n", p.Name)
	for _, s := range p.Steps {
		r := state.Get(s.Name)
		if r == nil {
			continue
		}
		fmt.Fprintf(os.Stderr, "  %-24s %s", s.Name, r.Status)
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, " (%v)", r.Err)
		}
		if verbose && len(r.Records) > 0 {
			fmt.Fprintf(os.Stderr, " [%d records]", len(r.Records))
		}
		fmt.Fprintln(os.Stderr)
	}
}
