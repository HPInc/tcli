// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/hpinc/tcli/pkg/env"
)

// main is the entry point for the tcli application.
func main() {
	if err := env.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tcli:", err)
		os.Exit(1)
	}
}
