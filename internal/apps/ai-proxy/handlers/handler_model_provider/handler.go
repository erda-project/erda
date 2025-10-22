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

package handler_model_provider

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ModelProviderHandler struct {
	DAO dao.DAO
}

func (h *ModelProviderHandler) Create(ctx context.Context, req *pb.ModelProviderCreateRequest) (*pb.ModelProvider, error) {
	return h.DAO.ModelProviderClient().Create(ctx, req)
}

func (h *ModelProviderHandler) Get(ctx context.Context, req *pb.ModelProviderGetRequest) (*pb.ModelProvider, error) {
	resp, err := h.DAO.ModelProviderClient().Get(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	desensitizeProvider(ctx, resp)
	return resp, nil
}

func (h *ModelProviderHandler) Delete(ctx context.Context, req *pb.ModelProviderDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ModelProviderClient().Delete(ctx, req)
}

func (h *ModelProviderHandler) Update(ctx context.Context, req *pb.ModelProviderUpdateRequest) (*pb.ModelProvider, error) {
	resp, err := h.DAO.ModelProviderClient().Update(ctx, req)
	if err != nil {
		return nil, err
	}
	desensitizeProvider(ctx, resp)
	return resp, nil
}

func (h *ModelProviderHandler) Paging(ctx context.Context, req *pb.ModelProviderPagingRequest) (*pb.ModelProviderPagingResponse, error) {
	resp, err := h.DAO.ModelProviderClient().Paging(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	for _, item := range resp.List {
		desensitizeProvider(ctx, item)
	}
	return resp, nil
}

func desensitizeProvider(ctx context.Context, item *pb.ModelProvider) {
	// pass for: admin
	if auth.IsAdmin(ctx) {
		return
	}
	// if a provider is client-belonged, pass for the client
	if item.ClientId != "" && auth.IsClient(ctx) {
		return
	}
	// hide sensitive data
	item.ApiKey = ""
	item.Metadata.Secret = nil
}
