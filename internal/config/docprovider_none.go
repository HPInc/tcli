// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"net/http"
)

type DocProviderNone struct {
}

func (d *DocProviderNone) Init() error {
	return nil
}

func (d *DocProviderNone) HttpRequest(req *http.Request) {
	if !logger.IsVerbose() {
		return
	}

	logger.Println("Request details:")
	pl := logger.GetPlainLogger()

	for name, values := range req.Header {
		for _, value := range values {
			pl.Printf("%s: %s\n", name, value)
		}
	}
	pl.Println("Method:", req.Method)
	pl.Println("Url:", req.URL)
	if body := baseGetBody(req); body != "" {
		pl.Println("Body:", body)
	}
}

func (d *DocProviderNone) HttpResponse(resp *http.Response, body []byte) {
	if !logger.IsVerbose() {
		return
	}
	baseHttpResponse(resp, body)
}
