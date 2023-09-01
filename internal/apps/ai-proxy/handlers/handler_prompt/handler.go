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

package handler_prompt

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type PromptHandler struct {
	DAO dao.DAO
}

func (h *PromptHandler) Create(ctx context.Context, req *pb.PromptCreateRequest) (*pb.Prompt, error) {
	return h.DAO.PromptClient().Create(ctx, req)
}

func (h *PromptHandler) Get(ctx context.Context, req *pb.PromptGetRequest) (*pb.Prompt, error) {
	return h.DAO.PromptClient().Get(ctx, req)
}

func (h *PromptHandler) Delete(ctx context.Context, req *pb.PromptDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.PromptClient().Delete(ctx, req)
}

func (h *PromptHandler) Update(ctx context.Context, req *pb.PromptUpdateRequest) (*pb.Prompt, error) {
	return h.DAO.PromptClient().Update(ctx, req)
}

func (h *PromptHandler) Paging(ctx context.Context, req *pb.PromptPagingRequest) (*pb.PromptPagingResponse, error) {
	return h.DAO.PromptClient().Paging(ctx, req)
}
