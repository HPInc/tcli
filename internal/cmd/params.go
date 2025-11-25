// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hpinc/tcli/internal/common"
	"github.com/hpinc/tcli/internal/config"
	"github.com/hpinc/tcli/internal/parser"
)

const JwtParam = "jwt"
const StatusParam = "status_code"
const GlobalFlags = "global"

type GlobalResult struct {
	Count       uint
	RetryCount  int64
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

// New creates a new ParseResult and initializes the flag set
// with the parameters from the given method and global flags.
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

// getParamVal returns the value of the parameter from the input map
// if it exists, otherwise it returns the default value of the parameter.
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

// getGlobal initializes the global flags and returns a GlobalResult.
func getGlobal(fs *flag.FlagSet, r *parser.Root) *GlobalResult {
	g := GlobalResult{}
	fs.StringVar(&g.BasePath, "base_path", getBasePath(r), "http base path")
	fs.StringVar(&g.Doc, "doc", "none", "Generate docs (none, shell)")
	fs.StringVar(&g.Format, "format", "", "json format")
	fs.Int64Var(&g.RetryCount, "retry_count", 10, "Number of retries on failure")
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

// ValidateParams checks that all required parameters have values
// and sets global configuration options based on the parsed values.
func (p *ParseResult) ValidateParams() error {
	for _, v := range p.Method.Parameters {
		val, ok := p.Values[v.Name]
		if !ok || val == nil || *val == "" {
			if v.Required {
				log.Errorf("Error: No value for required param: %s\n", v.Name)
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

// getParamValues processes the parameters and returns a paramsResult
// containing the URL values, path, body, and headers for the HTTP request.
func (p *ParseResult) getParamValues() (*paramsResult, error) {
	result := paramsResult{
		urlValues: &url.Values{},
		path:      p.Path,
		body:      "",
		headers:   make(http.Header),
	}
	for _, v := range p.Method.Parameters {
		val, ok := p.Values[v.Name]
		if !ok || val == nil || *val == "" {
			log.Debugf("Skipping empty parameter: %v\n", v.Name)
			continue
		}
		if v.In == parser.InPath {
			match := fmt.Sprintf("{%s}", v.Name)
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
