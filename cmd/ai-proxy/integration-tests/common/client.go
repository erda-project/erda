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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

// Client AI Proxy HTTP client
type Client struct {
	httpClient *http.Client
	config     *config.Config
}

// NewClient creates a new client
func NewClient() *Client {
	cfg := config.Get()
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
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

// GetWithHeaders sends GET request with custom headers
func (c *Client) GetWithHeaders(ctx context.Context, path string, headers map[string]string) *APIResponse {
	return c.sendRequestWithHeaders(ctx, "GET", path, nil, headers)
}

// Delete sends DELETE request
func (c *Client) Delete(ctx context.Context, path string) *APIResponse {
	return c.sendRequest(ctx, "DELETE", path, nil)
}

// DeleteWithHeaders sends DELETE request with custom headers
func (c *Client) DeleteWithHeaders(ctx context.Context, path string, headers map[string]string) *APIResponse {
	return c.sendRequestWithHeaders(ctx, "DELETE", path, nil, headers)
}

// PostJSONStreamWithHeaders sends JSON POST request with custom headers and handles streaming response
func (c *Client) PostJSONStreamWithHeaders(ctx context.Context, path string, payload interface{}, headers map[string]string, callback func(data []byte) error) *APIResponse {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("marshal request: %w", err)}
	}

	url := c.config.Host + path
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("create request: %w", err)}
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	fmt.Printf("→ %s %s\n", req.Method, req.URL.String())
	fmt.Printf("→ Headers: %+v\n", req.Header)
	fmt.Printf("→ Body: %s\n", string(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("send request: %w", err)}
	}
	defer resp.Body.Close()

	fmt.Printf("← %s %d\n", req.URL.String(), resp.StatusCode)
	fmt.Printf("← Headers: %+v\n", resp.Header)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       body,
			Error:      fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Handle streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse Server-Sent Events format
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			if callback != nil {
				if err := callback([]byte(data)); err != nil {
					return &APIResponse{Error: fmt.Errorf("callback error: %w", err)}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return &APIResponse{Error: fmt.Errorf("scan response: %w", err)}
	}

	return &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}
}

// PostJSONStream sends JSON POST request and handles streaming response
func (c *Client) PostJSONStream(ctx context.Context, path string, payload interface{}, callback func(data []byte) error) *APIResponse {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("marshal request: %w", err)}
	}

	url := c.config.Host + path
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("create request: %w", err)}
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}
	//req.Header.Set("X-Request-ID", "test-request-id") // Test request ID

	fmt.Printf("→ %s %s\n", req.Method, req.URL.String())
	fmt.Printf("→ Headers: %+v\n", req.Header)
	fmt.Printf("→ Body: %s\n", string(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("send request: %w", err)}
	}
	defer resp.Body.Close()

	fmt.Printf("← %s %d\n", req.URL.String(), resp.StatusCode)
	fmt.Printf("← Headers: %+v\n", resp.Header)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       body,
			Error:      fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Handle streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse Server-Sent Events format
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			if callback != nil {
				if err := callback([]byte(data)); err != nil {
					return &APIResponse{Error: fmt.Errorf("callback error: %w", err)}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return &APIResponse{Error: fmt.Errorf("scan response: %w", err)}
	}

	return &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}
}

// sendRequest sends request
func (c *Client) sendRequest(ctx context.Context, method, path string, payload interface{}) *APIResponse {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return &APIResponse{Error: fmt.Errorf("marshal request: %w", err)}
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := c.config.Host + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("create request: %w", err)}
	}

	// Set request headers
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}
	//req.Header.Set("X-Request-ID", "test-request-id") // Test request ID

	fmt.Printf("→ %s %s\n", req.Method, req.URL.String())
	fmt.Printf("→ Headers: %+v\n", req.Header)
	if payload != nil {
		jsonData, _ := json.Marshal(payload)
		fmt.Printf("→ Body: %s\n", string(jsonData))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("send request: %w", err)}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("read response: %w", err)}
	}

	fmt.Printf("← %s %d\n", req.URL.String(), resp.StatusCode)
	fmt.Printf("← Headers: %+v\n", resp.Header)
	fmt.Printf("← Body: %s\n", string(responseBody))

	result := &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}

	if resp.StatusCode >= 400 {
		result.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	return result
}

// sendRequestWithHeaders sends request with custom headers
func (c *Client) sendRequestWithHeaders(ctx context.Context, method, path string, payload interface{}, headers map[string]string) *APIResponse {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return &APIResponse{Error: fmt.Errorf("marshal request: %w", err)}
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := c.config.Host + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("create request: %w", err)}
	}

	// Set request headers
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	fmt.Printf("→ %s %s\n", req.Method, req.URL.String())
	fmt.Printf("→ Headers: %+v\n", req.Header)
	if payload != nil {
		jsonData, _ := json.Marshal(payload)
		fmt.Printf("→ Body: %s\n", string(jsonData))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("send request: %w", err)}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("read response: %w", err)}
	}

	fmt.Printf("← %s %d\n", req.URL.String(), resp.StatusCode)
	fmt.Printf("← Headers: %+v\n", resp.Header)
	fmt.Printf("← Body: %s\n", string(responseBody))

	result := &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}

	if resp.StatusCode >= 400 {
		result.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	return result
}

// PostJSONWithHeaders sends JSON request with custom headers
func (c *Client) PostJSONWithHeaders(ctx context.Context, path string, payload interface{}, headers map[string]string) *APIResponse {
	return c.sendRequestWithHeaders(ctx, "POST", path, payload, headers)
}

// PostMultipartWithHeaders sends multipart/form-data request with custom headers
func (c *Client) PostMultipartWithHeaders(ctx context.Context, path string, body io.Reader, contentType string, headers map[string]string) *APIResponse {
	url := c.config.Host + path
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("create request: %w", err)}
	}

	// Set request headers
	req.Header.Set("Content-Type", contentType)
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	fmt.Printf("→ %s %s\n", req.Method, req.URL.String())
	fmt.Printf("→ Headers: %+v\n", req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("send request: %w", err)}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("read response: %w", err)}
	}

	fmt.Printf("← %s %d\n", req.URL.String(), resp.StatusCode)
	fmt.Printf("← Headers: %+v\n", resp.Header)
	fmt.Printf("← Body: %s\n", string(responseBody))

	result := &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}

	if resp.StatusCode >= 400 {
		result.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	return result
}

// PostMultipart sends multipart/form-data request
func (c *Client) PostMultipart(ctx context.Context, path string, body io.Reader, contentType string) *APIResponse {
	url := c.config.Host + path
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("create request: %w", err)}
	}

	// Set request headers
	req.Header.Set("Content-Type", contentType)
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	fmt.Printf("→ %s %s\n", req.Method, req.URL.String())
	fmt.Printf("→ Headers: %+v\n", req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("send request: %w", err)}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{Error: fmt.Errorf("read response: %w", err)}
	}

	fmt.Printf("← %s %d\n", req.URL.String(), resp.StatusCode)
	fmt.Printf("← Headers: %+v\n", resp.Header)
	fmt.Printf("← Body: %s\n", string(responseBody))

	result := &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}

	if resp.StatusCode >= 400 {
		result.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	return result
}
