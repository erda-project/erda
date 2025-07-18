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

package handler_client_model_relation

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ClientModelRelationHandler struct {
	DAO dao.DAO
}

func (h *ClientModelRelationHandler) Allocate(ctx context.Context, req *pb.AllocateRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ClientModelRelationClient().Allocate(ctx, req)
}

func (h *ClientModelRelationHandler) UnAllocate(ctx context.Context, req *pb.AllocateRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ClientModelRelationClient().UnAllocate(ctx, req)
}

func (h *ClientModelRelationHandler) ListClientModels(ctx context.Context, req *pb.ListClientModelsRequest) (*pb.ListAllocatedModelsResponse, error) {
	return h.DAO.ClientModelRelationClient().ListClientModels(ctx, req)
}

func (h *ClientModelRelationHandler) Paging(ctx context.Context, req *pb.PagingRequest) (*pb.PagingResponse, error) {
	return h.DAO.ClientModelRelationClient().Paging(ctx, req)
}
