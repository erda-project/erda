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
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ModelHandler struct {
	DAO dao.DAO
}

func (h *ModelHandler) Create(ctx context.Context, req *pb.ModelCreateRequest) (*pb.Model, error) {
	return h.DAO.ModelClient().Create(ctx, req)
}

func (h *ModelHandler) Get(ctx context.Context, req *pb.ModelGetRequest) (*pb.Model, error) {
	return h.DAO.ModelClient().Get(ctx, req)
}

func (h *ModelHandler) Update(ctx context.Context, req *pb.ModelUpdateRequest) (*pb.Model, error) {
	return h.DAO.ModelClient().Update(ctx, req)
}

func (h *ModelHandler) Delete(ctx context.Context, req *pb.ModelDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ModelClient().Delete(ctx, req)
}
