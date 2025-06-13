// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package common

import "errors"

var (
	ErrNotFound             = errors.New("Not found")
	ErrMissingRequiredParam = errors.New("Required param missing value")
	ErrMoreChoices          = errors.New("There are additional choices")
	ErrNoSuchModule         = errors.New("Module not found")
	ErrNoSuchCommand        = errors.New("Command not found")
	ErrNoSuchSubCommand     = errors.New("Subcommand not found")
)
