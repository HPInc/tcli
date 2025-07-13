// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"path/filepath"
)

// apply the current selected profile
func (c *Config) applyProfile() {
	p := c.getProfileByName(c.ProfileName)
	if p == nil {
		logger.Fatalf("profile %s not found", c.ProfileName)
	}
	c.CurrentProfile = p
}

// look up profile by name
func (c *Config) getProfileByName(profile string) *Profile {
	for _, p := range c.Profiles {
		if p.Name == profile {
			return &p
		}
	}
	return nil
}

// operate on profile
func (p *Profile) getModuleConfig() string {
	return filepath.Join(getConfigRoot(), ModuleFileName)
}
