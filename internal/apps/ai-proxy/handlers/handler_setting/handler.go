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

package handler_setting

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/setting/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type Handler struct {
	DAO dao.DAO
}

func (h *Handler) Create(ctx context.Context, req *pb.SettingCreateRequest) (*pb.Setting, error) {
	return h.DAO.SettingClient().Create(ctx, req)
}

func (h *Handler) Get(ctx context.Context, req *pb.SettingGetRequest) (*pb.Setting, error) {
	return h.DAO.SettingClient().Get(ctx, req)
}

func (h *Handler) Delete(ctx context.Context, req *pb.SettingDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.SettingClient().Delete(ctx, req)
}

func (h *Handler) Update(ctx context.Context, req *pb.SettingUpdateRequest) (*pb.Setting, error) {
	return h.DAO.SettingClient().Update(ctx, req)
}

func (h *Handler) Paging(ctx context.Context, req *pb.SettingPagingRequest) (*pb.SettingPagingResponse, error) {
	return h.DAO.SettingClient().Paging(ctx, req)
}
