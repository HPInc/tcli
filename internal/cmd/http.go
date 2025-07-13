// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"openapi/internal/utils"
	"strconv"
)

const IgnoreError = -1

type HttpCommand struct {
	CmdBase
}

func init() {
	cmds["http"] = &HttpCommand{}
}

func (c *HttpCommand) Init(p *ParseResult) Command {
	k := HttpCommand{}
	k.baseInit(p, k.http)
	return &k
}

func (c *HttpCommand) http() error {
	p := c.p
	result, err := p.getParamValues()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s://%s%s%s", p.Global.Scheme, p.Global.Server,
		p.Global.BasePath, result.path)

	req, err := http.NewRequest(
		p.Method.MethodName,
		url,
		getData(result.body))
	if err != nil {
		log.Errorf("Request error: %v\n", err)
	}
	req.Header = result.headers
	if len(*result.urlValues) > 0 {
		req.URL.RawQuery = result.urlValues.Encode()
	}
	for _, m := range p.Method.Consumes {
		req.Header.Add("Content-Type", m)
	}
	// assumes oauth2 and populates authorization header
	// if security is reported.
	if _, ok := p.Values[JwtParam]; ok {
		utils.AddAuthorizationHeader(req, *p.Values[JwtParam])
	}

	client := utils.RetriableClient(p.Global.RetryCount)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Response error: %v\n", err)
	}
	// there might be useful results even when there is an error
	// so show the result first
	err = showResult(resp, p)
	if err != nil {
		return err
	}
	// check if the status code matches expected
	// this is normally 200 but can be overridden
	code, err := strconv.Atoi(*p.Values[StatusParam])
	if err != nil {
		code = 200
	}
	if resp.StatusCode != code {
		log.Errorf("StatusCode: %d, Expected: %d",
			resp.StatusCode, code)
		if !p.Global.IgnoreError {
			log.Fatalf("Exit on first error")
		}
	}
	return nil
}

func showResult(r *http.Response, p *ParseResult) error {
	if r == nil || r.Body == nil {
		return nil
	}
	defer utils.CloseBodyReader(r.Body)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Read error: %v\n", err)
		return err
	}
	log.HttpResponse(r, body)
	applyFormat(body, p)
	return nil
}

func getData(body string) io.Reader {
	if body != "" {
		return bytes.NewBuffer([]byte(body))
	}
	return nil
}

func applyFormat(bytes []byte, p *ParseResult) {
	if p.Global.Format != "" {
		utils.DoFormat(bytes, p.Values, p.Global.Format)
	} else {
		fmt.Println(string(bytes))
	}
}
