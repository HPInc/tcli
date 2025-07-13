// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"openapi/internal/common"
	"openapi/internal/config"
	"openapi/internal/parser"
)

const JwtParam = "jwt"
const StatusParam = "status_code"
const GlobalFlags = "global"

type GlobalResult struct {
	Count       uint
	RetryCount  uint
	BasePath    string
	Scheme      string
	Server      string
	Doc         string
	Format      string
	Verbose     bool
	Parallel    bool
	IgnoreError bool
}

type ParseResult struct {
	Path   string
	Flags  *flag.FlagSet
	Values common.Values
	Method *parser.Method
	Global *GlobalResult
}

type paramsResult struct {
	urlValues *url.Values
	path      string
	body      string
	headers   http.Header
}

func New(o *parser.Method, in *common.Input, r *parser.Root) *ParseResult {
	fs := flag.NewFlagSet(o.OperationId, flag.ExitOnError)

	p := ParseResult{
		Flags:  fs,
		Path:   o.Path,
		Values: make(map[string]*string),
		Method: o,
		Global: getGlobal(fs, r),
	}

	for _, i := range o.Parameters {
		d := i.GetDescription()
		p.Values[i.Name] = fs.String(i.Name, getParamVal(&i, in), d)
	}
	if o.NeedJwt(r) {
		p.Values[JwtParam] = fs.String(JwtParam,
			config.GetTokenCache().GetToken(), "bearer token")
	}
	p.Values[StatusParam] = fs.String(StatusParam, "200",
		"Status code to check")
	return &p
}

func getParamVal(p *parser.Parameter, i *common.Input) string {
	if i != nil {
		if input, ok := (*i)[p.Name]; ok {
			if p.In == parser.InBody {
				return common.GetJsonString(input)
			}
			return fmt.Sprintf("%v", input)
		}
	}
	return p.DefaultStr()
}

func getGlobal(fs *flag.FlagSet, r *parser.Root) *GlobalResult {
	g := GlobalResult{}
	fs.StringVar(&g.BasePath, "base_path", getBasePath(r), "http base path")
	fs.StringVar(&g.Doc, "doc", "none", "Generate docs (none, shell)")
	fs.StringVar(&g.Format, "format", "", "json format")
	fs.UintVar(&g.RetryCount, "retry_count", 10, "Number of retries on failure")
	fs.StringVar(&g.Scheme, "scheme", getScheme(r), "Scheme")
	fs.StringVar(&g.Server, "server", getHost(r), "Server")
	fs.UintVar(&g.Count, "count", 1, "Number of times to repeat command")
	fs.BoolVar(&g.IgnoreError, "ignore_errors", false, "Ignore errors")
	fs.BoolVar(&g.Parallel, "parallel", false, "Do runs in parallel")
	fs.BoolVar(&g.Verbose, "v", false, "Verbose")
	return &g
}

// potential config point for a per module
// default from modules.yaml
func getHost(r *parser.Root) string {
	if r.Host == "" {
		return "localhost:8080"
	} else {
		return r.Host
	}
}

// potential config point for a per module
// default from modules.yaml
func getBasePath(r *parser.Root) string {
	return r.BasePath
}

// potential config point for a per module
// default from modules.yaml
func getScheme(r *parser.Root) string {
	if len(r.Schemes) == 0 {
		return "http"
	} else {
		return r.Schemes[0]
	}
}

func (p *ParseResult) ValidateParams() error {
	for _, v := range p.Method.Parameters {
		val, ok := p.Values[v.Name]
		if !ok || val == nil || *val == "" {
			log.Errorf("Error: No value for required param: %s\n", v.Name)
			if v.Required {
				return common.ErrMissingRequiredParam
			} else {
				continue
			}
		}
	}
	if p.Global.Verbose {
		config.SetVerboseLogging()
	}
	config.SetDocType(p.Global.Doc)
	return nil
}

func (p *ParseResult) getParamValues() (*paramsResult, error) {
	result := paramsResult{
		urlValues: &url.Values{},
		path:      p.Path,
		body:      "",
		headers:   make(http.Header),
	}
	for _, v := range p.Method.Parameters {
		match := fmt.Sprintf("{%s}", v.Name)
		val, ok := p.Values[v.Name]
		if !ok || val == nil {
			continue
		}
		if v.In == parser.InPath {
			result.path = strings.Replace(result.path, match, *val, 1)
		} else if v.In == parser.InQuery {
			result.urlValues.Set(v.Name, fmt.Sprintf("%s", *val))
		} else if v.In == parser.InBody {
			result.body = *val
		} else if v.In == parser.InHeader {
			// skip authorization header for better handling with jwt oauth path
			if strings.ToLower(v.Name) == "authorization" {
				continue
			}
			result.headers.Add(v.Name, *val)
		}
	}
	return &result, nil
}
