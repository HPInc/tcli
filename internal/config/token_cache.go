// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"openapi/internal/common"
	"strings"
)

const (
	TokenFile = ".token"
)

var (
	// cache file path
	TokenCacheFile = fmt.Sprintf("%s/%s", getConfigDir(), TokenFile)
)

type TokenCache struct {
	tokenData string
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		tokenData: readCachedData(),
	}
}

// set access token
func (e *TokenCache) Update(bytes []byte) error {
	if err := common.WriteFile(TokenCacheFile, bytes); err != nil {
		return err
	}
	return nil
}

// read cached data if any
func readCachedData() string {
	bytes, err := common.ReadFile(TokenCacheFile)
	if err != nil {
		logger.Debugf("Error reading access token from cache: file=%s, err=%v",
			TokenCacheFile, err)
		return ""
	}
	return strings.TrimSpace(string(bytes))
}

// get token data
func (e *TokenCache) GetToken() string {
	return e.tokenData
}
