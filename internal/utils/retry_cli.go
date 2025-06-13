// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// wraps retry logic in overrides to http.Client
package utils

import (
	"io"
	"net/http"
	"openapi/internal/config"
	"strconv"
	"time"
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
	MaxRetry   uint
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
func RetriableClient(retries uint) *Client {
	c := NewClient()
	c.Do = c.RetriableDo
	c.MaxRetry = retries
	c.StatusCode = 0
	return c
}

func RetriableClientWithStatus(retries uint, status int) *Client {
	c := RetriableClient(retries)
	c.StatusCode = status
	return c
}

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

func (c *Client) RetriableDo(req *http.Request) (*http.Response, error) {
	logger := config.GetLogger()
	logger.HttpRequest(req)
	var i uint
	for {
		resp, err := c.HttpClient.Do(req)
		if c.RetryWithBackoff(logger, i, resp, err) {
			logger.Debugf("%d/%d\n", i, c.MaxRetry)
			i++
		} else {
			return resp, err
		}
	}
}

func (c *Client) RetryWithBackoff(logger *config.Log,
	i uint, resp *http.Response, err error) bool {
	doRetry := false
	retrySeconds := time.Second * time.Duration(i)
	if err != nil {
		logger.Errorf("error = %v, retriable.\n", err)
		doRetry = true
	} else if resp.StatusCode == 0 || resp.StatusCode > 500 {
		logger.Debugf("status = %d, retriable.\n", resp.StatusCode)
		doRetry = true
	} else if resp.StatusCode == 429 {
		doRetry = true
		retrySeconds = time.Duration(getRetryAfterHeaderValue(resp))
		logger.Debugf(
			"status = %d, retry_after = %d seconds, retriable.\n",
			resp.StatusCode, retrySeconds)
	} else if c.StatusCode > 0 && resp.StatusCode != c.StatusCode {
		// allow for an override to wait for a specific code
		doRetry = true
		logger.Debugf("status = %d, retrying till %d.\n", resp.StatusCode, c.StatusCode)
	}
	if doRetry && i < c.MaxRetry {
		time.Sleep(retrySeconds)
		return true
	}
	return false
}

func getRetryAfterHeaderValue(resp *http.Response) time.Duration {
	retryAfter := DefaultRetryAfterSeconds
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
