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

package handler_mcp_server

import (
	"bytes"
	"context"
	"errors"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"regexp"
	"strconv"
	"strings"
)

var addrRegex = regexp.MustCompile(`^https?://([^:/]+)(?::(\d+))?$`)

type MCPHandler struct {
	DAO               dao.DAO
	McpProxyPublicURL string
}

func NewMCPHandler(d dao.DAO, addr string) *MCPHandler {
	handler := MCPHandler{
		DAO:               d,
		McpProxyPublicURL: strings.TrimSuffix(addr, "/"),
	}
	return &handler
}

func (m *MCPHandler) Update(ctx context.Context, req *pb.MCPServerUpdateRequest) (*pb.MCPServerUpdateResponse, error) {
	return m.DAO.MCPServerClient().Update(ctx, req)
}

func (m *MCPHandler) Delete(ctx context.Context, req *pb.MCPServerDeleteRequest) (*pb.MCPServerDeleteResponse, error) {
	return m.DAO.MCPServerClient().Delete(ctx, req)
}

func (m *MCPHandler) Version(ctx context.Context, req *pb.MCPServerVersionRequest) (*pb.MCPServerVersionResponse, error) {
	total, servers, err := m.DAO.MCPServerClient().List(ctx, &mcp_server.ListOptions{
		PageNum:            int(req.PageNum),
		PageSize:           int(req.PageSize),
		Name:               req.Name,
		IncludeUnpublished: req.IncludeUnpublished,
	})
	if err != nil {
		return nil, err
	}

	if !req.RawEndpoint {
		if err := VerifyAddr(m.McpProxyPublicURL); err != nil {
			return nil, err
		}

		for i := range servers {
			servers[i].Endpoint = m.buildEndpoint(servers[i])
		}
	}

	return &pb.MCPServerVersionResponse{
		Total: total,
		List:  servers,
	}, nil
}

func (m *MCPHandler) Register(ctx context.Context, req *pb.MCPServerRegisterRequest) (*pb.MCPServerRegisterResponse, error) {
	return m.DAO.MCPServerClient().CreateOrUpdate(ctx, req)
}

func (m *MCPHandler) Publish(ctx context.Context, req *pb.MCPServerActionPublishRequest) (*pb.MCPServerActionPublishResponse, error) {
	return m.DAO.MCPServerClient().Publish(ctx, req)
}

func (m *MCPHandler) Get(ctx context.Context, req *pb.MCPServerGetRequest) (*pb.MCPServerGetResponse, error) {
	resp, err := m.DAO.MCPServerClient().Get(ctx, req)
	mcpServer := resp.GetData()

	if !req.RawEndpoint {
		if err := VerifyAddr(m.McpProxyPublicURL); err != nil {
			return nil, err
		}
		mcpServer.Endpoint = m.buildEndpoint(mcpServer)
	}

	return resp, err
}

func (m *MCPHandler) List(ctx context.Context, req *pb.MCPServerListRequest) (*pb.MCPServerListResponse, error) {
	total, servers, err := m.DAO.MCPServerClient().List(ctx, &mcp_server.ListOptions{
		IncludeUnpublished: req.IncludeUnpublished,
		PageNum:            int(req.PageNum),
		PageSize:           int(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	if !req.RawEndpoint {
		if err := VerifyAddr(m.McpProxyPublicURL); err != nil {
			return nil, err
		}

		for i := range servers {
			servers[i].Endpoint = m.buildEndpoint(servers[i])
		}
	}

	return &pb.MCPServerListResponse{
		Total: total,
		List:  servers,
	}, nil
}

func (m *MCPHandler) buildEndpoint(server *pb.MCPServer) string {
	// http://127.0.0.1:8081/proxy/connect/demo/1.0.0
	buffer := bytes.Buffer{}
	buffer.WriteString(m.McpProxyPublicURL)
	buffer.WriteString("/proxy/connect/")
	buffer.WriteString(server.Name)
	buffer.WriteString("/")
	buffer.WriteString(server.Version)
	return buffer.String()
}

func VerifyAddr(addr string) error {
	if addr == "" {
		return errors.New("mcp proxy addr is empty")
	}

	matches := addrRegex.FindStringSubmatch(addr)
	if matches == nil {
		return errors.New("mcp proxy addr is invalid")
	}

	// matches[1] 是 host，matches[2] 是 port（可选）
	if matches[2] != "" {
		port, err := strconv.Atoi(matches[2])
		if err != nil || port < 1 || port > 65535 {
			return errors.New("mcp proxy addr port is invalid")
		}
	}

	return nil
}
