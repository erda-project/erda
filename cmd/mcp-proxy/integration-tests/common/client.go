// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/erda-project/erda/cmd/mcp-proxy/integration-tests/config"
)

// Client MCP Proxy HTTP client
type Client struct {
	HttpClient *http.Client
	Config     *config.Config
}

// NewClient creates a new client
func NewClient() *Client {
	cfg := config.Get()
	return &Client{
		HttpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		Config: cfg,
	}
}

// APIResponse generic API response
type APIResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Error      error
}

// IsSuccess checks if response is successful
func (r *APIResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// GetJSON parses JSON response
func (r *APIResponse) GetJSON(v interface{}) error {
	if r.Error != nil {
		return r.Error
	}
	return json.Unmarshal(r.Body, v)
}

// PostJSON sends JSON POST request
func (c *Client) PostJSON(ctx context.Context, path string, payload interface{}) *APIResponse {
	return c.sendRequest(ctx, "POST", path, payload)
}

// Get sends GET request
func (c *Client) Get(ctx context.Context, path string) *APIResponse {
	return c.sendRequest(ctx, "GET", path, nil)
}

// PutJSON sends JSON PUT request
func (c *Client) PutJSON(ctx context.Context, path string, payload interface{}) *APIResponse {
	return c.sendRequest(ctx, "PUT", path, payload)
}

// Delete sends DELETE request
func (c *Client) Delete(ctx context.Context, path string) *APIResponse {
	return c.sendRequest(ctx, "DELETE", path, nil)
}

// sendRequest sends HTTP request with JSON payload
func (c *Client) sendRequest(ctx context.Context, method, path string, payload interface{}) *APIResponse {
	var body io.Reader
	var err error

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return &APIResponse{Error: fmt.Errorf("marshal request: %w", err)}
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := c.Config.Host + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("create request: %w", err)}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.Config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Config.Token)
	}

	// Send request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("send request: %w", err)}
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("read response: %w", err)}
	}

	return &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}
}
