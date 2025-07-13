// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"net/http"
)

type DocProviderShell struct {
}

func (d *DocProviderShell) Init() error {
	return nil
}

// print curl request
func (d *DocProviderShell) HttpRequest(req *http.Request) {
	logger.Println("curl command:")
	pl := logger.GetPlainLogger()

	//method eg: -X GET
	method := fmt.Sprintf("-X %s", req.Method)

	//proxy eg: -x <server:port>
	proxy := ""
	proxyUrl, err := http.ProxyFromEnvironment(req)
	if err == nil && proxyUrl != nil {
		proxy = fmt.Sprintf("-x %s", proxyUrl)
	}

	//headers eg: -H "Content-Type: application/json"
	headers := ""
	for name, values := range req.Header {
		for _, value := range values {
			header := fmt.Sprintf(`-H "%s: %s" `, name, value)
			headers = headers + header
		}
	}

	//body
	body := baseGetBody(req)
	if body != "" {
		body = fmt.Sprintf(`-d '%s'`, body)
	}
	pl.Printf("curl %s %s %s %s %v\n",
		method, proxy, headers, body, req.URL)
}

func (d *DocProviderShell) HttpResponse(resp *http.Response, body []byte) {
	baseHttpResponse(resp, body)
}
