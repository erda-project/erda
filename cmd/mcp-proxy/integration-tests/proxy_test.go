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

package integration_tests

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"testing"
	"time"

	"github.com/erda-project/erda/cmd/mcp-proxy/integration-tests/config"
	"github.com/mark3labs/mcp-go/client/transport"
)

// ProxyConnectRequest represents the proxy connect request
type ProxyConnectRequest struct {
	MCPName string `json:"mcpName"`
	MCPTag  string `json:"mcpTag"`
}

// ProxyMessageRequest represents the proxy message request
type ProxyMessageRequest struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Err     *ErrorInfo  `json:"err,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestProxyConnectGET(t *testing.T) {
	cfg := config.Get()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	startTime := time.Now()

	sseClient, err := client.NewSSEMCPClient(cfg.Host+fmt.Sprintf("/proxy/connect/%s/%s", cfg.TestMCPName, cfg.TestMCPTag), transport.WithHeaders(map[string]string{
		"Authorization": cfg.Token,
	}))
	if err != nil {
		t.Errorf("create SSEMCPClient failed: %v", err)
	}

	if err = sseClient.Start(ctx); err != nil {
		t.Errorf("start SSEMCPClient failed: %v", err)
	}

	request := mcp.InitializeRequest{}
	request.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	request.Params.ClientInfo = mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}

	_, err = sseClient.Initialize(ctx, request)
	if err != nil {
		t.Errorf("init SSEMCPClient failed: %v", err)
	}

	result, err := sseClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: cfg.TestToolName,
			Arguments: map[string]string{
				cfg.TestArgsName: cfg.TestArgsValue,
			},
		},
	})
	if err != nil {
		t.Errorf("call SSEMCPClient failed: %v", err)
	}

	t.Logf("result: %+v", result)

	// Log response for debugging
	t.Logf("init SSEMCPClient duration: %v", time.Since(startTime))
}
