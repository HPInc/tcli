// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package utils

import (
	"time"

	"github.com/hpinc/tcli/pkg/config"
)

// RetryWait retries a function up to 'count' times, waiting an increasing amount of time between each attempt.
// The wait time increases linearly (1s, 2s, 3s, ...).
// If the function returns true, it stops retrying and returns true.
// If all attempts fail, it returns false.
func RetryWait(count int64, fn func() bool) bool {
	for i := int64(1); i <= count; i++ {
		if fn() {
			return true
		}
		config.GetLogger().Debugf("Waiting %d seconds..", i)
		time.Sleep(time.Duration(i) * time.Second)
	}
	return false
}
