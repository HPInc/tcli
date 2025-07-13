// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package utils

import (
	"fmt"
	"io"
	"net/http"
	"openapi/internal/config"
)

const (
	Authorization = "Authorization"
	ContentType   = "Content-Type"
	UserAgent     = "User-Agent"

	CliUserAgent = "tcli"
)

// generic add header
func AddRequestHeader(req *http.Request, key, value string) {
	req.Header.Add(key, value)
}

// add user-agent to http request
func AddUserAgentHeader(req *http.Request) {
	req.Header.Add(UserAgent, CliUserAgent)
}

// add content type application/json
func AddContentTypeJson(req *http.Request) {
	req.Header.Add(ContentType, "application/json")
}

// add content type url encoded
func AddContentTypeUrlEncoded(req *http.Request) {
	req.Header.Add(ContentType, "application/x-www-form-urlencoded")
}

// add authorization header to http request
func AddAuthorizationHeader(req *http.Request, bearerToken string) {
	if bearerToken != "" {
		req.Header.Add(Authorization,
			fmt.Sprintf("Bearer %s", bearerToken))
	}
}

// close w error check
func CloseBodyReader(rc io.ReadCloser) {
	if rc != nil {
		if err := rc.Close(); err != nil {
			config.GetLogger().Error(err)
		}
	}
}
