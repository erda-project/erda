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

package handler_mcp_server_template

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server_template/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server_config_instance"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type MCPTemplateHandler struct {
	dao    dao.DAO
	logger logs.Logger
}

func NewMcpTemplateHandler(dao dao.DAO, l logs.Logger) *MCPTemplateHandler {
	return &MCPTemplateHandler{dao: dao, logger: l}
}

func (m *MCPTemplateHandler) Get(ctx context.Context, request *pb.MCPServerTemplateGetRequest) (*pb.MCPServerTemplateGetResponse, error) {
	var clientId = ""

	if !auth.IsAdmin(ctx) {
		id, ok := ctxhelper.GetClientId(ctx)
		if !ok {
			m.logger.Error("client id should not be empty for non-admin")
			return nil, errors.New("failed to get clientId")
		}
		clientId = id
	}

	template, err := m.dao.MCPServerTemplateClient().Get(ctx, request.McpName, request.Version)
	if err != nil {
		logrus.Errorf("failed to get mcp server template info, err: %v", err)
		return nil, err
	}

	count, err := m.dao.MCPServerConfigInstanceClient().Count(ctx, request.McpName, request.Version, clientId)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			m.logger.Errorf("failed to get mcp server config instance info, err: %v", err)
			return nil, err
		}
	}

	if count == 0 && template.IsEmptyTemplate() && !auth.IsAdmin(ctx) {
		_, err = m.dao.MCPServerConfigInstanceClient().CreateOrUpdate(ctx, &mcp_server_config_instance.McpServerConfigInstance{
			McpName:  template.McpName,
			Version:  template.Version,
			Config:   "{}",
			ClientID: clientId,
		})
		if err != nil {
			m.logger.Errorf("failed to create mcp server config instance info, err: %v", err)
			return nil, err
		}
		count++
	}

	protobuf := template.ToProtobuf()
	protobuf.InstanceCount = count
	return &pb.MCPServerTemplateGetResponse{Template: protobuf.Template}, nil
}

func (m *MCPTemplateHandler) List(ctx context.Context, request *pb.MCPServerTemplateListRequest) (*pb.MCPServerTemplateListResponse, error) {
	var clientId = ""

	if !auth.IsAdmin(ctx) {
		id, ok := ctxhelper.GetClientId(ctx)
		if !ok {
			m.logger.Error("client id should not be empty for non-admin")
			return nil, errors.New("failed to get clientId")
		}
		clientId = id
	}

	list, total, err := m.dao.MCPServerTemplateClient().List(ctx, request.PageSize, request.PageNum)
	if err != nil {
		m.logger.Errorf("failed to get mcp server template list, err: %v", err)
		return nil, err
	}

	instances, _, err := m.dao.MCPServerConfigInstanceClient().List(ctx, &mcp_server_config_instance.ListOptions{PageSize: 0, PageNum: 0, ClientId: &clientId})
	if err != nil {
		m.logger.Errorf("failed to get mcp server config instance list, err: %v", err)
		return nil, err
	}

	var exist = make(map[string]int64)
	for _, instance := range instances {
		key := fmt.Sprintf("%s%s%s", instance.McpName, instance.Version, instance.ClientID)
		exist[key]++
	}

	var items = make([]*pb.MCPServerTemplateItem, 0, len(list))
	for _, template := range list {
		temp := template.ToProtobuf()

		key := fmt.Sprintf("%s%s%s", temp.McpName, temp.Version, clientId)
		temp.InstanceCount = exist[key]
		// if template is not nil, it means it can be configured
		temp.IsConfigurable = temp.Template != nil && len(temp.Template) != 0

		// If the instance does not exist, create an instance for the MCP with a template of {} or ''.
		if exist[key] == 0 && template.IsEmptyTemplate() && !auth.IsAdmin(ctx) {
			_, err := m.dao.MCPServerConfigInstanceClient().CreateOrUpdate(ctx, &mcp_server_config_instance.McpServerConfigInstance{
				McpName:  temp.McpName,
				Version:  temp.Version,
				Config:   "{}",
				ClientID: clientId,
			})
			if err != nil {
				m.logger.Errorf("failed to create mcp server config instance, err: %v", err)
				continue
			}
			temp.InstanceCount++
		}

		items = append(items, temp)
	}

	// sort templates
	slices.SortFunc(items, func(a, b *pb.MCPServerTemplateItem) int {
		if a.IsConfigurable {
			return -1
		}
		return 1
	})

	return &pb.MCPServerTemplateListResponse{
		List:  items,
		Total: uint64(total),
	}, nil
}

func (m *MCPTemplateHandler) Create(ctx context.Context, request *pb.MCPServerTemplateCreateRequest) (*pb.MCPServerTemplateCreateResponse, error) {
	created, err := m.dao.MCPServerTemplateClient().Create(ctx, request.Template, request.McpName, request.Version)
	if err != nil {
		logrus.Errorf("failed to create mcp server template, err: %v", err)
		return nil, err
	}
	return &pb.MCPServerTemplateCreateResponse{Data: created.ToProtobuf()}, nil
}
