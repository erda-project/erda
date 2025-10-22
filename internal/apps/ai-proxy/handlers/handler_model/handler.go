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

package handler_model

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ModelHandler struct {
	DAO dao.DAO
}

func (h *ModelHandler) Create(ctx context.Context, req *pb.ModelCreateRequest) (*pb.Model, error) {
	return h.DAO.ModelClient().Create(ctx, req)
}

func (h *ModelHandler) Get(ctx context.Context, req *pb.ModelGetRequest) (*pb.Model, error) {
	resp, err := h.DAO.ModelClient().Get(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	desensitizeModel(ctx, resp)
	return resp, nil
}

func (h *ModelHandler) Update(ctx context.Context, req *pb.ModelUpdateRequest) (*pb.Model, error) {
	resp, err := h.DAO.ModelClient().Update(ctx, req)
	if err != nil {
		return nil, err
	}
	desensitizeModel(ctx, resp)
	return resp, nil
}

func (h *ModelHandler) Delete(ctx context.Context, req *pb.ModelDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ModelClient().Delete(ctx, req)
}

func (h *ModelHandler) Paging(ctx context.Context, req *pb.ModelPagingRequest) (*pb.ModelPagingResponse, error) {
	resp, err := h.DAO.ModelClient().Paging(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	for _, item := range resp.List {
		desensitizeModel(ctx, item)
	}
	return resp, nil
}

func (h *ModelHandler) UpdateModelAbilitiesInfo(ctx context.Context, req *pb.ModelAbilitiesInfoUpdateRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ModelClient().UpdateModelAbilitiesInfo(ctx, req)
}

func desensitizeModel(ctx context.Context, item *pb.Model) {
	// pass for: admin
	if auth.IsAdmin(ctx) {
		return
	}
	// if a model is client-belonged, pass for the client
	if item.ClientId != "" && auth.IsClient(ctx) {
		return
	}
	// hide sensitive data for non-admin
	item.ApiKey = ""
	item.Metadata.Secret = nil
}
