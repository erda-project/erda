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

package handler_client_token

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ClientTokenHandler struct {
	DAO dao.DAO
}

func (h *ClientTokenHandler) Create(ctx context.Context, req *pb.ClientTokenCreateRequest) (*pb.ClientToken, error) {
	return h.DAO.ClientTokenClient().Create(ctx, req)
}

func (h *ClientTokenHandler) Get(ctx context.Context, req *pb.ClientTokenGetRequest) (*pb.ClientToken, error) {
	return h.DAO.ClientTokenClient().Get(ctx, req)
}

func (h *ClientTokenHandler) Delete(ctx context.Context, req *pb.ClientTokenDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ClientTokenClient().Delete(ctx, req)
}

func (h *ClientTokenHandler) Update(ctx context.Context, req *pb.ClientTokenUpdateRequest) (*pb.ClientToken, error) {
	return h.DAO.ClientTokenClient().Update(ctx, req)
}

func (h *ClientTokenHandler) Paging(ctx context.Context, req *pb.ClientTokenPagingRequest) (*pb.ClientTokenPagingResponse, error) {
	return h.DAO.ClientTokenClient().Paging(ctx, req)
}
