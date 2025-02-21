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

package handler_rich_client

import (
	"context"
	"fmt"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/client/rich_client/pb"
	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_model_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/pkg/strutil"
)

type ClientHandler struct {
	DAO dao.DAO
}

func (h *ClientHandler) GetByAccessKeyId(ctx context.Context, req *pb.GetByClientAccessKeyIdRequest) (*pb.RichClient, error) {
	// check access key id
	if req.AccessKeyId == "" {
		if auth.IsClient(ctx) { // use client's access key id
			req.AccessKeyId = auth.GetClient(ctx).AccessKeyId
		}
	}
	if req.AccessKeyId == "" {
		return nil, fmt.Errorf("access key id required")
	}

	// check data permission
	if auth.IsClient(ctx) {
		authedClient := auth.GetClient(ctx)
		if authedClient.AccessKeyId != req.AccessKeyId {
			return nil, fmt.Errorf("you can only query your own client info")
		}
	}

	// get client
	clientPagingResp, err := h.DAO.ClientClient().Paging(ctx, &clientpb.ClientPagingRequest{PageNum: 1, PageSize: 1, AccessKeyIds: []string{req.AccessKeyId}})
	if err != nil {
		return nil, err
	}
	if clientPagingResp.Total != 1 {
		return nil, nil
	}
	client := clientPagingResp.List[0]

	// get models
	clientModelResp, err := h.DAO.ClientModelRelationClient().ListClientModels(ctx, &clientmodelrelationpb.ListClientModelsRequest{ClientId: client.Id})
	if err != nil {
		return nil, err
	}
	modelPagingResp, err := h.DAO.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
		PageNum:  1,
		PageSize: 999,
		Ids:      clientModelResp.ModelIds,
	})
	if err != nil {
		return nil, err
	}

	// get model providers
	var providerIds []string
	for _, model := range modelPagingResp.List {
		providerIds = append(providerIds, model.ProviderId)
	}
	providerIds = strutil.DedupSlice(providerIds, true)
	providerPagingResp, err := h.DAO.ModelProviderClient().Paging(ctx, &modelproviderpb.ModelProviderPagingRequest{
		PageNum:  1,
		PageSize: 999,
		Ids:      providerIds,
	})
	if err != nil {
		return nil, err
	}
	providerMapById := make(map[string]*modelproviderpb.ModelProvider)
	for _, provider := range providerPagingResp.List {
		providerMapById[provider.Id] = handler_model_provider.Handle_for_display(provider)
	}

	// assign rich models
	richModels := make([]*pb.RichModel, 0, len(modelPagingResp.List))
	for _, model := range modelPagingResp.List {
		richModels = append(richModels, &pb.RichModel{
			Model:    model,
			Provider: providerMapById[model.ProviderId],
		})
	}
	// assign rich client
	richClient := &pb.RichClient{
		Client: client,
		Models: richModels,
	}

	return richClient, nil
}
