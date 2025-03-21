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

package handler_mcp

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp-server/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type MCPHandler struct {
	DAO dao.DAO
}

func (m *MCPHandler) Register(ctx context.Context, req *pb.MCPServerRegisterRequest) (*pb.MCPServerRegisterResponse, error) {
	return m.DAO.MCPServerClient().CreateOrUpdate(ctx, req)
}

func (m *MCPHandler) Publish(ctx context.Context, req *pb.MCPServerActionPublishRequest) (*pb.MCPServerActionPublishResponse, error) {
	return m.DAO.MCPServerClient().Publish(ctx, req)
}

func (m *MCPHandler) Get(ctx context.Context, req *pb.MCPServerGetRequest) (*pb.MCPServerGetResponse, error) {
	return m.DAO.MCPServerClient().Get(ctx, req)
}

func (m *MCPHandler) List(ctx context.Context, req *pb.MCPServerListRequest) (*pb.MCPServerListResponse, error) {
	return m.DAO.MCPServerClient().Paging(ctx, req)
}
