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
	"context"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"

	cmspb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_mcp_relation/pb"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
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
	var servers []*pb.MCPServer
	var total int64

	var scopes = make(map[string]*cmspb.ScopeIdList)

	if !auth.IsAdmin(ctx) && !auth.IsInternalClient(ctx) {
		clientId, ok := ctxhelper.GetClientId(ctx)
		if !ok {
			return nil, errors.New("clientId not found")
		}

		result, err := m.DAO.ClientMCPRelationClient().ListClientMCPScope(ctx, &cmspb.ListClientMCPScopeRequest{
			ClientId: clientId,
		})
		if err != nil {
			return nil, err
		}

		scopes = result.Scope
		scopes["*"] = &cmspb.ScopeIdList{Ids: []string{"0"}}
		scopes["client"] = &cmspb.ScopeIdList{Ids: []string{clientId}}
	} else {
		scopes["*"] = &cmspb.ScopeIdList{Ids: []string{}}
	}

	for scopeType, scope := range scopes {
		count, temp, err := m.DAO.MCPServerClient().List(ctx, &mcp_server.ListOptions{
			PageNum:            int(req.PageNum),
			PageSize:           int(req.PageSize),
			Name:               req.Name,
			IncludeUnpublished: req.IncludeUnpublished,
			ScopeIds:           scope.Ids,
			ScopeType:          scopeType,
		})
		if err != nil {
			return nil, err
		}

		total += count
		servers = append(servers, temp...)
	}

	if !req.UseRawEndpoint {
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
	var scopes = make(map[string]*cmspb.ScopeIdList)

	if !auth.IsAdmin(ctx) && !auth.IsInternalClient(ctx) {
		clientId, ok := ctxhelper.GetClientId(ctx)
		if !ok {
			return nil, errors.New("clientId not found")
		}

		result, err := m.DAO.ClientMCPRelationClient().ListClientMCPScope(ctx, &cmspb.ListClientMCPScopeRequest{
			ClientId: clientId,
		})
		if err != nil {
			return nil, err
		}

		scopes = result.Scope
		scopes["*"] = &cmspb.ScopeIdList{Ids: []string{"0"}}
		scopes["client"] = &cmspb.ScopeIdList{Ids: []string{clientId}}
	} else {
		scopes["*"] = &cmspb.ScopeIdList{Ids: []string{}}
	}

	resp, err := m.DAO.MCPServerClient().Get(ctx, req)
	mcpServer := resp.GetData()

	if scopes["*"] == nil && !slices.Contains(scopes[mcpServer.ScopeType].Ids, mcpServer.ScopeId) {
		return nil, errors.New("no permission to access")
	}

	if !req.UseRawEndpoint {
		if err := VerifyAddr(m.McpProxyPublicURL); err != nil {
			return nil, err
		}
		mcpServer.Endpoint = m.buildEndpoint(mcpServer)
	}

	return resp, err
}

func (m *MCPHandler) List(ctx context.Context, req *pb.MCPServerListRequest) (*pb.MCPServerListResponse, error) {
	var servers []*pb.MCPServer
	var total int64

	var scopes = make(map[string]*cmspb.ScopeIdList)

	if !auth.IsAdmin(ctx) && !auth.IsInternalClient(ctx) {
		clientId, ok := ctxhelper.GetClientId(ctx)
		if !ok {
			return nil, errors.New("clientId not found")
		}

		result, err := m.DAO.ClientMCPRelationClient().ListClientMCPScope(ctx, &cmspb.ListClientMCPScopeRequest{
			ClientId: clientId,
		})
		if err != nil {
			return nil, err
		}

		scopes = result.Scope
		// add default scope
		scopes["*"] = &cmspb.ScopeIdList{Ids: []string{"0"}}
		scopes["client"] = &cmspb.ScopeIdList{Ids: []string{clientId}}
	} else {
		scopes["*"] = &cmspb.ScopeIdList{Ids: []string{}}
	}

	for scopeType, scope := range scopes {
		count, temp, err := m.DAO.MCPServerClient().List(ctx, &mcp_server.ListOptions{
			IncludeUnpublished: req.IncludeUnpublished,
			PageNum:            int(req.PageNum),
			PageSize:           int(req.PageSize),
			ScopeIds:           scope.Ids,
			ScopeType:          scopeType,
		})
		if err != nil {
			return nil, err
		}

		total += count
		servers = append(servers, temp...)
	}

	if !req.UseRawEndpoint {
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
	return m.McpProxyPublicURL + "/proxy/connect/" + server.Name + "/" + server.Version
}

func VerifyAddr(addr string) error {
	if addr == "" {
		return errors.New("mcp proxy addr is empty")
	}

	matches := addrRegex.FindStringSubmatch(addr)
	if matches == nil {
		return errors.New("mcp proxy addr is invalid")
	}

	// matches[1] is host, matches[2] is port（optional）
	if matches[2] != "" {
		port, err := strconv.Atoi(matches[2])
		if err != nil || port < 1 || port > 65535 {
			return errors.New("mcp proxy addr port is invalid")
		}
	}

	return nil
}
