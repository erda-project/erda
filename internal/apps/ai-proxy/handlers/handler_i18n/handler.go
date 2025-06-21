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

package handler_i18n

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/i18n/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type I18nHandler struct {
	DAO dao.DAO
}

func (h *I18nHandler) Create(ctx context.Context, req *pb.I18NCreateRequest) (*pb.I18NConfig, error) {
	return h.DAO.I18nClient().Create(ctx, req)
}

func (h *I18nHandler) Get(ctx context.Context, req *pb.I18NGetRequest) (*pb.I18NConfig, error) {
	return h.DAO.I18nClient().Get(ctx, req)
}

func (h *I18nHandler) Delete(ctx context.Context, req *pb.I18NDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.I18nClient().Delete(ctx, req)
}

func (h *I18nHandler) Update(ctx context.Context, req *pb.I18NUpdateRequest) (*pb.I18NConfig, error) {
	return h.DAO.I18nClient().Update(ctx, req)
}

func (h *I18nHandler) Paging(ctx context.Context, req *pb.I18NPagingRequest) (*pb.I18NPagingResponse, error) {
	return h.DAO.I18nClient().Paging(ctx, req)
}

func (h *I18nHandler) BatchCreate(ctx context.Context, req *pb.I18NBatchCreateRequest) (*pb.I18NBatchCreateResponse, error) {
	return h.DAO.I18nClient().BatchCreate(ctx, req)
}

func (h *I18nHandler) GetByConfig(ctx context.Context, req *pb.I18NGetByConfigRequest) (*pb.I18NConfig, error) {
	return h.DAO.I18nClient().GetByConfig(ctx, req)
}
