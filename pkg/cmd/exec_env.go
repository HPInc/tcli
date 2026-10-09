// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"sync"

	"github.com/hpinc/tcli/pkg/common"
	"github.com/hpinc/tcli/pkg/config"
	"github.com/hpinc/tcli/pkg/parser"
)

// Environment runs commands and fails fast: the first error stops new
// jobs and input records from starting. Jobs already running finish
// their current attempt but do not retry. Wait reports every error.
type Environment struct {
	Args     []string
	Parallel *ParallelEnvironment

	mu       sync.Mutex
	errs     []error
	stop     chan struct{} // closed on the first error
	stopOnce sync.Once
}

func GetExecutionEnv(args []string) *Environment {
	return &Environment{
		Args: args,
		stop: make(chan struct{}),
	}
}

// Exec runs the method for one input record. Errors are recorded and
// reported by Wait as well as returned.
func (e *Environment) Exec(m *parser.Method, i *common.Input) error {
	c, err := e.getCmd(m, i)
	if err != nil {
		e.addError(err)
		return err
	}
	// this is here because we do not have early indication
	// on parallel exec env till we parse cmd line args
	if c.global.Parallel {
		if e.Parallel == nil {
			e.Parallel = newParallel(e.addError, e.Stopped)
		}
	}
	for i := uint(1); i <= c.global.Count; i++ {
		e.addJob(c)
	}
	return nil
}

// Wait blocks until all jobs finish and returns every job error joined,
// or nil if all jobs succeeded.
func (e *Environment) Wait() error {
	if e.Parallel != nil {
		e.Parallel.wait()
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return errors.Join(e.errs...)
}

// Stopped reports whether a job has failed, meaning no further jobs or
// input records should be started.
func (e *Environment) Stopped() bool {
	select {
	case <-e.stop:
		return true
	default:
		return false
	}
}

func (e *Environment) addJob(c *CmdBase) {
	if e.Stopped() {
		return
	}
	if e.Parallel != nil {
		e.Parallel.addJob(c)
	} else {
		e.addError(c.runFunc())
	}
}

func (e *Environment) addError(err error) {
	if err == nil {
		return
	}
	e.mu.Lock()
	e.errs = append(e.errs, err)
	e.mu.Unlock()
	e.stopOnce.Do(func() { close(e.stop) })
}

func (e *Environment) getCmd(m *parser.Method, i *common.Input) (*CmdBase, error) {
	p := New(m, i, config.GetCurrentModule().ConfigRoot)
	if err := p.Flags.Parse(e.Args); err != nil {
		return nil, err
	}
	if err := p.ValidateParams(); err != nil {
		return nil, err
	}
	// if there is an extension, it points to an implementation
	// such as sqs or mqtt to handle
	c, err := GetCommand(m.GetExtensionClass())
	if err != nil {
		return nil, err
	}
	b := c.Init(p).GetBase()
	b.stop = e.stop
	return b, nil
}
