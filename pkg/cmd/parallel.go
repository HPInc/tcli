// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"runtime"
	"sync"
)

type ParallelEnvironment struct {
	numWorkers int
	wg         sync.WaitGroup
	jobs       chan *CmdBase
	onError    func(error)
	// skip reports whether queued jobs should be dropped (fail fast)
	skip func() bool
}

func newParallel(onError func(error), skip func() bool) *ParallelEnvironment {
	pe := &ParallelEnvironment{
		numWorkers: runtime.NumCPU(),
		jobs:       make(chan *CmdBase),
		onError:    onError,
		skip:       skip,
	}
	for w := 1; w <= pe.numWorkers; w++ {
		go pe.worker(w)
	}
	return pe
}

func (pe *ParallelEnvironment) wait() {
	close(pe.jobs)
	pe.wg.Wait()
}

func (pe *ParallelEnvironment) worker(w int) {
	log.Debug("Starting worker:", w)
	for j := range pe.jobs {
		if pe.skip() {
			log.Debugf("Worker %d: skipping job after failure\n", w)
		} else {
			log.Debugf("Worker %d: starting job\n", w)
			pe.onError(j.runFunc())
			log.Debugf("Worker %d: ending job\n", w)
		}
		pe.wg.Done()
	}
}

func (pe *ParallelEnvironment) addJob(c *CmdBase) {
	pe.wg.Add(1)
	pe.jobs <- c
}
