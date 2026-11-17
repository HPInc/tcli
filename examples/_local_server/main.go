// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// petstore_local starts the in-memory Petstore used by the integration tests.
package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/hpinc/tcli/internal/petstoretest"
)

func main() {
	server := petstoretest.NewServer()
	defer server.Close()
	fmt.Printf("Petstore listening at %s/v2\n", server.URL)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
