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
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ModelProviderHandler struct {
	DAO dao.DAO
}

func (h *ModelProviderHandler) Create(ctx context.Context, req *pb.ModelProviderCreateRequest) (*pb.ModelProvider, error) {
	return h.DAO.ModelProviderClient().Create(ctx, req)
}

func (h *ModelProviderHandler) Get(ctx context.Context, req *pb.ModelProviderGetRequest) (*pb.ModelProvider, error) {
	return h.DAO.ModelProviderClient().Get(ctx, req)
}

func (h *ModelProviderHandler) Delete(ctx context.Context, req *pb.ModelProviderDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ModelProviderClient().Delete(ctx, req)
}

func (h *ModelProviderHandler) Update(ctx context.Context, req *pb.ModelProviderUpdateRequest) (*pb.ModelProvider, error) {
	return h.DAO.ModelProviderClient().Update(ctx, req)
}

func (h *ModelProviderHandler) Paging(ctx context.Context, req *pb.ModelProviderPagingRequest) (*pb.ModelProviderPagingResponse, error) {
	pagingResult, err := h.DAO.ModelProviderClient().Paging(ctx, req)
	if err != nil {
		return nil, err
	}
	// handle for display
	for _, provider := range pagingResult.List {
		Handle_for_display(provider)
	}
	return pagingResult, nil
}

func Handle_for_display(p *pb.ModelProvider) *pb.ModelProvider {
	if p.Metadata == nil || p.Metadata.Public == nil {
		return p
	}
	if displayProviderType := p.Metadata.Public["displayProviderType"]; displayProviderType != "" {
		p.Type = displayProviderType
	}
	return p
}
