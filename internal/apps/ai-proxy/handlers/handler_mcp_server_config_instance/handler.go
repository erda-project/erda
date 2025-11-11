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

package handler_mcp_server_config_instance

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server_config_instance/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server_config_instance"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type MCPConfigInstanceHandler struct {
	dao    dao.DAO
	logger logs.Logger
}

func NewMCPConfigInstanceHandler(dao dao.DAO, logger logs.Logger) *MCPConfigInstanceHandler {
	return &MCPConfigInstanceHandler{dao: dao, logger: logger}
}

func (m *MCPConfigInstanceHandler) Get(ctx context.Context, request *pb.MCPServerConfigInstanceGetRequest) (*pb.MCPServerConfigInstanceGetResponse, error) {
	clientId, ok := ctxhelper.GetClientId(ctx)
	if (!ok && !auth.IsAdmin(ctx)) || (auth.IsAdmin(ctx) && request.ClientId == nil) {
		m.logger.Errorf("client id or client-id should not be empty")
		return nil, errors.New("failed to get clientId")
	} else if auth.IsAdmin(ctx) {
		clientId = *request.ClientId
	}

	instance, err := m.dao.MCPServerConfigInstanceClient().Get(ctx, request.McpName, request.Version, clientId)
	if err != nil {
		m.logger.Errorf("failed to get instance: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pb.MCPServerConfigInstanceGetResponse{
				Data: &pb.MCPServerConfigInstanceItem{
					Version: request.Version,
					McpName: request.McpName,
					Config:  make(map[string]*structpb.Value),
				},
			}, nil
		}

		return nil, err
	}
	return &pb.MCPServerConfigInstanceGetResponse{
		Data: instance.ToProtobuf(),
	}, nil
}

func (m *MCPConfigInstanceHandler) List(ctx context.Context, request *pb.MCPServerConfigInstanceListRequest) (*pb.MCPServerConfigInstanceListResponse, error) {
	var clientId *string
	if !auth.IsAdmin(ctx) {
		if id, ok := ctxhelper.GetClientId(ctx); ok {
			clientId = &id
		} else {
			m.logger.Errorf("client id or client-id should not be empty")
			return nil, errors.New("failed to get clientId")
		}
	}

	list, total, err := m.dao.MCPServerConfigInstanceClient().List(ctx, &mcp_server_config_instance.ListOptions{
		ClientId: clientId,
		PageNum:  int(request.PageNum),
		PageSize: int(request.PageSize),
	})
	if err != nil {
		m.logger.Errorf("failed to list instance: %v", err)
		return nil, err
	}
	var instances = make([]*pb.MCPServerConfigInstanceItem, 0, len(list))
	for _, instance := range list {
		instances = append(instances, instance.ToProtobuf())
	}

	return &pb.MCPServerConfigInstanceListResponse{
		List:  instances,
		Total: uint64(total),
	}, nil
}

func (m *MCPConfigInstanceHandler) Create(ctx context.Context, request *pb.MCPServerConfigInstanceCreateRequest) (*pb.MCPServerConfigInstanceCreateResponse, error) {
	var clientId = ""

	if auth.IsAdmin(ctx) {
		if request.ClientId == nil {
			m.logger.Error("client id should not be empty for admin")
			return nil, errors.New("failed to get clientId")
		}
		clientId = *request.ClientId
	} else {
		id, ok := ctxhelper.GetClientId(ctx)
		if !ok {
			m.logger.Error("client id should not be empty for non-admin")
			return nil, errors.New("failed to get clientId")
		}
		clientId = id
	}

	config, err := json.Marshal(request.Config)
	if err != nil {
		m.logger.Errorf("failed to marshal config: %v", err)
		return nil, err
	}

	instance, err := m.dao.MCPServerConfigInstanceClient().CreateOrUpdate(ctx, &mcp_server_config_instance.McpServerConfigInstance{
		McpName:  request.McpName,
		Version:  request.Version,
		Config:   string(config),
		ClientID: clientId,
	})
	if err != nil {
		m.logger.Errorf("failed to create instance: %v", err)
		return nil, err
	}
	return &pb.MCPServerConfigInstanceCreateResponse{Data: instance.ToProtobuf()}, nil
}

func (m *MCPConfigInstanceHandler) Update(ctx context.Context, request *pb.MCPServerConfigInstanceUpdateRequest) (*commonpb.VoidResponse, error) {
	var clientId = ""
	if auth.IsAdmin(ctx) {
		if request.ClientId == nil {
			m.logger.Error("client id should not be empty for admin")
			return nil, errors.New("failed to get clientId")
		}
		clientId = *request.ClientId
	} else {
		id, ok := ctxhelper.GetClientId(ctx)
		if !ok {
			m.logger.Error("client id should not be empty for non-admin")
			return nil, errors.New("failed to get clientId")
		}
		clientId = id
	}

	config, err := json.Marshal(request.Config)
	if err != nil {
		m.logger.Errorf("failed to marshal config: %v", err)
		return nil, err
	}

	if err := m.dao.MCPServerConfigInstanceClient().UpdateConfig(ctx, request.McpName, request.Version, clientId, string(config)); err != nil {
		m.logger.Errorf("failed to update instance: %v", err)
		return nil, err
	}
	return new(commonpb.VoidResponse), nil
}

func (m *MCPConfigInstanceHandler) Delete(ctx context.Context, request *pb.MCPServerConfigInstanceDeleteRequest) (*commonpb.VoidResponse, error) {
	var clientId = ""
	if auth.IsAdmin(ctx) {
		if request.ClientId == nil {
			m.logger.Error("client id should not be empty for admin")
			return nil, errors.New("failed to get clientId")
		}
		clientId = *request.ClientId
	} else {
		id, ok := ctxhelper.GetClientId(ctx)
		if !ok {
			m.logger.Error("client id should not be empty for non-admin")
			return nil, errors.New("failed to get clientId")
		}
		clientId = id
	}

	if err := m.dao.MCPServerConfigInstanceClient().Delete(ctx, request.McpName, request.Version, clientId); err != nil {
		m.logger.Errorf("failed to delete instance: %v", err)
		return nil, err
	}
	return new(commonpb.VoidResponse), nil
}
