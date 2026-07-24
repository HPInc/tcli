// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// wraps retry logic in overrides to http.Client
package utils

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/hpinc/tcli/pkg/config"
)

const (
	RetryAfter               = "Retry-After"
	DefaultRetryAfterSeconds = time.Second * 3
	DefaultRetryCount        = 3
)

type Do func(req *http.Request) (*http.Response, error)

type Client struct {
	HttpClient *http.Client
	Do         Do
	MaxRetry   int64
	StatusCode int
}

// no friction wrapper
func NewClient() *Client {
	return &Client{
		HttpClient: http.DefaultClient,
		Do:         http.DefaultClient.Do,
	}
}

// retry enabled client
func RetriableClient(retries int64) *Client {
	c := NewClient()
	c.Do = c.RetriableDo
	c.MaxRetry = retries
	c.StatusCode = 0
	return c
}

// retry enabled client with specific status code to wait for
// e.g. wait for 200 OK
func RetriableClientWithStatus(retries int64, status int) *Client {
	c := RetriableClient(retries)
	c.StatusCode = status
	return c
}

// get with retry
func (c *Client) Get(url string) (*http.Response, error) {
	return c.doRequest(url, "GET", nil)
}

// post with retry
func (c *Client) Post(url string, body io.Reader) (*http.Response, error) {
	return c.doRequest(url, "POST", body)
}

// handle requests
func (c *Client) doRequest(url, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	AddUserAgentHeader(req)
	config.GetLogger().HttpRequest(req)
	return c.Do(req)
}

// retry logic
func (c *Client) RetriableDo(req *http.Request) (*http.Response, error) {
	logger := config.GetLogger()
	logger.HttpRequest(req)
	var i int64
	for {
		if i > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			req.Body = body
		}
		// #nosec G704 -- request target is intentionally provided by CLI input.
		resp, err := c.HttpClient.Do(req)
		if c.RetryWithBackoff(logger, i, resp, err) {
			logger.Debugf("%d/%d\n", i, c.MaxRetry)
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			i++
		} else {
			return resp, err
		}
	}
}

// determine if we should retry, and backoff if so
// returns true if we should retry
// sleeps for backoff duration if retrying
func (c *Client) RetryWithBackoff(logger *config.Log,
	i int64, resp *http.Response, err error) bool {
	doRetry := false
	statusCode := 0
	retrySeconds := time.Second * time.Duration(i)
	if resp != nil {
		statusCode = resp.StatusCode
	}
	if err != nil {
		logger.Errorf("error = %v, retriable.\n", err)
		doRetry = true
	} else if statusCode == 0 || statusCode >= 500 {
		logger.Debugf("status = %d, retriable.\n", statusCode)
		doRetry = true
	} else if statusCode == 429 {
		doRetry = true
		retrySeconds = time.Duration(getRetryAfterHeaderValue(resp))
		logger.Debugf(
			"status = %d, retry_after = %d seconds, retriable.\n",
			statusCode, retrySeconds)
	} else if c.StatusCode > 0 && statusCode != c.StatusCode {
		// allow for an override to wait for a specific code
		doRetry = true
		logger.Debugf("status = %d, retrying till %d.\n", statusCode, c.StatusCode)
	}
	if doRetry && i < c.MaxRetry {
		time.Sleep(retrySeconds)
		return true
	}
	return false
}

// get the retry-after header value in seconds
// default to DefaultRetryAfterSeconds if not present or invalid
func getRetryAfterHeaderValue(resp *http.Response) time.Duration {
	retryAfter := DefaultRetryAfterSeconds
	if resp == nil {
		return retryAfter
	}
	var err error
	var i int
	if val, ok := resp.Header[RetryAfter]; ok {
		if i, err = strconv.Atoi(val[0]); err != nil {
			config.GetLogger().Errorf(
				"Error: invalid retry-after value: %v\n", val)
		}
		retryAfter = time.Second * time.Duration(i)
	}
	return retryAfter
}
