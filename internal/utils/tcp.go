// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package utils

import (
	"openapi/internal/config"
	"time"
)

func RetryWait(count uint, fn func() bool) bool {
	for i := uint(1); i <= count; i++ {
		if fn() {
			return true
		}
		config.GetLogger().Debugf("Waiting %d seconds..", i)
		time.Sleep(time.Duration(i) * time.Second)
	}
	return false
}
