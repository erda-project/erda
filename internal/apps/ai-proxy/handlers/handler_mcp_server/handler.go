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
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	cmspb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_mcp_relation/pb"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server_config_instance"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/common/apis"
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

	var clientId = ""

	if !auth.IsAdmin(ctx) {
		clientId = ctxhelper.MustGetClientId(ctx)
	}

	var servers []*pb.MCPServer
	var total int64

	var scopes = make([]mcp_server.Scope, 0)

	if !auth.IsAdmin(ctx) && apis.GetInternalClient(ctx) == "" {
		result, err := m.DAO.ClientMCPRelationClient().ListClientMCPScope(ctx, &cmspb.ListClientMCPScopeRequest{
			ClientId: clientId,
		})
		if err != nil {
			return nil, err
		}

		for tpy, scopeIds := range result.Scope {
			for _, id := range scopeIds.Ids {
				scopes = append(scopes, mcp_server.Scope{
					ScopeType: tpy,
					ScopeId:   id,
				})
			}
		}

		// add default scope
		scopes = append(scopes, mcp_server.Scope{
			ScopeType: vars.McpScopeTypePlatform,
			ScopeId:   vars.McpAnyScopeId,
		})

		scopes = append(scopes, mcp_server.Scope{
			ScopeType: vars.McpScopeTypeClientId,
			ScopeId:   clientId,
		})

	} else {
		scopes = append(scopes, mcp_server.Scope{
			ScopeType: vars.McpAnyScopeType,
			ScopeId:   vars.McpAnyScopeId,
		})
	}

	total, servers, err := m.DAO.MCPServerClient().List(ctx, &mcp_server.ListOptions{
		PageNum:            int(req.PageNum),
		PageSize:           int(req.PageSize),
		Name:               req.Name,
		IncludeUnpublished: req.IncludeUnpublished,
		Scopes:             scopes,
	})
	if err != nil {
		return nil, err
	}

	// select instance count
	instances, err := m.DAO.MCPServerConfigInstanceClient().CountAll(ctx, clientId)
	if err != nil {
		return nil, err
	}

	var count = make(map[string]int64)
	for _, instance := range instances {
		key := fmt.Sprintf("%s-%s", instance.McpName, instance.Version)
		count[key] = int64(instance.Count)
	}

	for i, server := range servers {
		if !req.UseRawEndpoint {
			if err := VerifyAddr(m.McpProxyPublicURL); err != nil {
				return nil, err
			}
			servers[i].Endpoint = m.buildEndpoint(servers[i])
		}
		if count[fmt.Sprintf("%s-%s", server.Name, server.Version)] == 0 {
			// if it has no instance, create it
			c, err := m.createInstance(ctx, server.Name, server.Version, clientId)
			if err != nil {
				return nil, err
			}
			count[fmt.Sprintf("%s-%s", server.Name, server.Version)] += c
		}
		servers[i].InstanceCount = count[fmt.Sprintf("%s-%s", servers[i].Name, servers[i].Version)]
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

	var clientId = ""

	if !auth.IsAdmin(ctx) {
		clientId = ctxhelper.MustGetClientId(ctx)
	}

	var scopes = make(map[string]*cmspb.ScopeIdList)

	if !auth.IsAdmin(ctx) && apis.GetInternalClient(ctx) == "" {
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
		scopes[vars.McpAnyScopeType] = &cmspb.ScopeIdList{Ids: []string{"0"}}
		scopes[vars.McpScopeTypeClientId] = &cmspb.ScopeIdList{Ids: []string{clientId}}
	} else {
		scopes[vars.McpAnyScopeType] = &cmspb.ScopeIdList{Ids: []string{}}
	}

	resp, err := m.DAO.MCPServerClient().Get(ctx, req)
	if err != nil {
		return nil, err
	}
	mcpServer := resp.GetData()

	if scopes["*"] == nil && !slices.Contains(scopes[mcpServer.ScopeType].Ids, mcpServer.ScopeId) {
		return nil, errors.New("no permission to access")
	}

	count, err := m.DAO.MCPServerConfigInstanceClient().Count(ctx, mcpServer.Name, mcpServer.Version, clientId)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		c, err := m.createInstance(ctx, mcpServer.Name, mcpServer.Version, clientId)
		if err != nil {
			return nil, err
		}
		count += c
	}

	mcpServer.InstanceCount = count

	if !req.UseRawEndpoint {
		if err := VerifyAddr(m.McpProxyPublicURL); err != nil {
			return nil, err
		}
		mcpServer.Endpoint = m.buildEndpoint(mcpServer)
	}

	return resp, err
}

func (m *MCPHandler) List(ctx context.Context, req *pb.MCPServerListRequest) (*pb.MCPServerListResponse, error) {

	var clientId = ""

	if !auth.IsAdmin(ctx) {
		clientId = ctxhelper.MustGetClientId(ctx)
	}

	var servers []*pb.MCPServer
	var total int64

	var scopes = make([]mcp_server.Scope, 0)

	if !auth.IsAdmin(ctx) && apis.GetInternalClient(ctx) == "" {

		result, err := m.DAO.ClientMCPRelationClient().ListClientMCPScope(ctx, &cmspb.ListClientMCPScopeRequest{
			ClientId: clientId,
		})
		if err != nil {
			return nil, err
		}
		for tpy, scopeIds := range result.Scope {
			for _, id := range scopeIds.Ids {
				scopes = append(scopes, mcp_server.Scope{
					ScopeType: tpy,
					ScopeId:   id,
				})
			}
		}

		// add default scope
		scopes = append(scopes, mcp_server.Scope{
			ScopeType: vars.McpScopeTypePlatform,
			ScopeId:   vars.McpAnyScopeId,
		})

		scopes = append(scopes, mcp_server.Scope{
			ScopeType: vars.McpScopeTypeClientId,
			ScopeId:   clientId,
		})

	} else {
		scopes = append(scopes, mcp_server.Scope{
			ScopeType: vars.McpAnyScopeType,
			ScopeId:   vars.McpAnyScopeId,
		})
	}

	total, servers, err := m.DAO.MCPServerClient().List(ctx, &mcp_server.ListOptions{
		IncludeUnpublished: req.IncludeUnpublished,
		PageNum:            int(req.PageNum),
		PageSize:           int(req.PageSize),
		Scopes:             scopes,
	})
	if err != nil {
		return nil, err
	}

	// select instance count
	instances, err := m.DAO.MCPServerConfigInstanceClient().CountAll(ctx, clientId)
	if err != nil {
		return nil, err
	}

	var count = make(map[string]int64)
	for _, instance := range instances {
		key := fmt.Sprintf("%s-%s", instance.McpName, instance.Version)
		count[key] = int64(instance.Count)
	}

	for i, server := range servers {
		if !req.UseRawEndpoint {
			if err := VerifyAddr(m.McpProxyPublicURL); err != nil {
				return nil, err
			}
			servers[i].Endpoint = m.buildEndpoint(servers[i])
		}
		if count[fmt.Sprintf("%s-%s", server.Name, server.Version)] == 0 {
			logrus.Errorf("server ====> %s/%s has no instance", server.Name, server.Version)
			// if it has no instance, create it
			c, err := m.createInstance(ctx, server.Name, server.Version, clientId)
			if err != nil {
				return nil, err
			}
			count[fmt.Sprintf("%s-%s", server.Name, server.Version)] += c
		}
		servers[i].InstanceCount = count[fmt.Sprintf("%s-%s", servers[i].Name, servers[i].Version)]
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

func (m *MCPHandler) createInstance(ctx context.Context, mcpServer string, version string, clientId string) (int64, error) {
	if clientId == "" {
		return 0, nil
	}
	template, err := m.DAO.MCPServerTemplateClient().Get(ctx, mcpServer, version)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if template.IsEmptyTemplate() {
		_, err = m.DAO.MCPServerConfigInstanceClient().CreateOrUpdate(ctx, &mcp_server_config_instance.McpServerConfigInstance{
			InstanceName: mcp_server_config_instance.DefaultInstanceName,
			Version:      version,
			ClientID:     clientId,
			McpName:      template.McpName,
			Config:       mcp_server_config_instance.EmptyConfig,
		})
		if err == nil {
			return 1, nil
		}
		logrus.Errorf("create server instance [%s] failed: %v", mcpServer, err)
	}

	return 0, nil
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
