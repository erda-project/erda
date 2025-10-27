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

package handler_service_provider

import (
	"context"
	"fmt"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ServiceProviderHandler struct {
	DAO dao.DAO
}

func (h *ServiceProviderHandler) Create(ctx context.Context, req *pb.ServiceProviderCreateRequest) (*pb.ServiceProvider, error) {
	return h.DAO.ServiceProviderClient().Create(ctx, req)
}

func (h *ServiceProviderHandler) Get(ctx context.Context, req *pb.ServiceProviderGetRequest) (*pb.ServiceProvider, error) {
	resp, err := h.DAO.ServiceProviderClient().Get(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	desensitizeProvider(ctx, resp)
	return resp, nil
}

func (h *ServiceProviderHandler) Delete(ctx context.Context, req *pb.ServiceProviderDeleteRequest) (*commonpb.VoidResponse, error) {
	// check models
	modelPagingResp, err := h.DAO.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
		PageNum:    1,
		PageSize:   1,
		ProviderId: req.Id,
	})
	if err != nil {
		return nil, err
	}
	if modelPagingResp.Total > 0 {
		return nil, fmt.Errorf("provider has related models, can not delete")
	}
	return h.DAO.ServiceProviderClient().Delete(ctx, req)
}

func (h *ServiceProviderHandler) Update(ctx context.Context, req *pb.ServiceProviderUpdateRequest) (*pb.ServiceProvider, error) {
	resp, err := h.DAO.ServiceProviderClient().Update(ctx, req)
	if err != nil {
		return nil, err
	}
	desensitizeProvider(ctx, resp)
	return resp, nil
}

func (h *ServiceProviderHandler) Paging(ctx context.Context, req *pb.ServiceProviderPagingRequest) (*pb.ServiceProviderPagingResponse, error) {
	resp, err := h.DAO.ServiceProviderClient().Paging(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	for _, item := range resp.List {
		desensitizeProvider(ctx, item)
	}
	return resp, nil
}

func desensitizeProvider(ctx context.Context, item *pb.ServiceProvider) {
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
