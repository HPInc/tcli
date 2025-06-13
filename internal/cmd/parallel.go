// Copyright 2025 HP Development Company, L.P.
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
}

func newParallel() *ParallelEnvironment {
	pe := &ParallelEnvironment{
		numWorkers: runtime.NumCPU(),
		jobs:       make(chan *CmdBase),
	}
	for w := 1; w <= pe.numWorkers; w++ {
		go worker(w, pe.jobs, &pe.wg)
	}
	return pe
}

func (pe *ParallelEnvironment) wait() {
	close(pe.jobs)
	pe.wg.Wait()
}

func worker(w int, jobs <-chan *CmdBase, wg *sync.WaitGroup) {
	log.Debug("Starting worker:", w)
	for j := range jobs {
		log.Debugf("Worker %d: starting job\n", w)
		_ = j.runFunc()
		wg.Done()
		log.Debugf("Worker %d: ending job\n", w)
	}
}

func (pe *ParallelEnvironment) addJob(c *CmdBase) {
	pe.wg.Add(1)
	pe.jobs <- c
}
