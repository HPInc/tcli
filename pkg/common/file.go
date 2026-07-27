// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package common

import (
	"os"
	"path/filepath"
)

const (
	WritePermissions = 0600
)

// read file and return bytes
// uses filepath.Clean for gosec error G304
func ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Clean(name))
}

// address gosec write permissions
// address gosec filename variable warnings
func WriteFile(name string, data []byte) error {
	return os.WriteFile(filepath.Clean(name), data, WritePermissions)
}

func OpenFile(name string) (*os.File, error) {
	return os.Open(filepath.Clean(name))
}
