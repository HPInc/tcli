// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package main

import (
	"os"

	"github.com/hpinc/tcli/pkg/env"
)

// main is the entry point for the tcli application.
func main() {
	if err := env.Run(); err != nil {
		os.Exit(1)
	}
}
