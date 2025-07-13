// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"io"
	"net/http"
)

// doctype declaration
type DocType string

const (
	None  DocType = ""
	Shell DocType = "shell"
)

// define interface for document generation
type DocProvider interface {
	// initialize, set a handler
	Init() error
	// provide docs for http request
	HttpRequest(*http.Request)
	// provide docs for http response
	HttpResponse(*http.Response, []byte)
}

func NewDocProvider(dt string) DocProvider {
	p := &DocProviderNone{}
	switch DocType(dt) {
	case None:
		return p
	case Shell:
		return &DocProviderShell{}
	default:
		return p
	}
}

// base response logging
func baseHttpResponse(resp *http.Response, body []byte) {
	logger.Println("Response details:")
	pl := logger.GetPlainLogger()
	pl.Println("Status:", resp.Status)
	for name, values := range resp.Header {
		for _, value := range values {
			pl.Printf("%s: %s\n", name, value)
		}
	}
}

// GetBody has some nuances. cover it here to not repeat
// everywhere
func baseGetBody(req *http.Request) string {
	if req.GetBody == nil {
		return ""
	}
	body := ""
	bodyReader, err := req.GetBody()
	if bodyReader != nil && err == nil {
		defer CloseBodyReader(bodyReader)
		result, err := io.ReadAll(bodyReader)
		if err == nil && len(result) > 0 {
			body = string(result)
		}
	}
	return body
}

// close w error check
func CloseBodyReader(rc io.ReadCloser) {
	if rc != nil {
		if err := rc.Close(); err != nil {
			logger.Error(err)
		}
	}
}
