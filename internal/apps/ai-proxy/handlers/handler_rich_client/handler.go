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
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n/i18n_services"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ClientHandler struct {
	DAO dao.DAO
}

func (h *ClientHandler) GetByAccessKeyId(ctx context.Context, req *pb.GetByClientAccessKeyIdRequest) (*pb.RichClient, error) {
	// set context
	ctxhelper.PutDBClient(ctx, h.DAO)

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
	client, ok := ctxhelper.GetClient(ctx)
	if !ok {
		clientPagingResp, err := h.DAO.ClientClient().Paging(ctx, &clientpb.ClientPagingRequest{PageNum: 1, PageSize: 1, AccessKeyIds: []string{req.AccessKeyId}})
		if err != nil {
			return nil, err
		}
		if clientPagingResp.Total != 1 {
			return nil, nil
		}
		client = clientPagingResp.List[0]
	}

	//// get models from two parts:
	//// - platform-models: assigned from client-model-relation
	//// - client-models: belonged client itself
	//var allModels []*modelpb.Model
	//// platform-models
	//clientModelResp, err := h.DAO.ClientModelRelationClient().ListClientModels(ctx, &clientmodelrelationpb.ListClientModelsRequest{ClientId: client.Id})
	//if err != nil {
	//	return nil, err
	//}
	//if len(clientModelResp.ModelIds) > 0 {
	//	platformModelPagingResp, err := h.DAO.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
	//		PageNum:  1,
	//		PageSize: 999,
	//		Ids:      clientModelResp.ModelIds,
	//	})
	//	if err != nil {
	//		return nil, err
	//	}
	//	allModels = append(allModels, platformModelPagingResp.List...)
	//}
	//// client-models
	//clientModelPagingResp, err := h.DAO.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
	//	PageNum:  1,
	//	PageSize: 999,
	//	ClientId: client.Id,
	//})
	//if err != nil {
	//	return nil, err
	//}
	//allModels = append(allModels, clientModelPagingResp.List...)
	//allModels = strutil.DedupAnySlice(allModels, func(i int) any {
	//	return allModels[i].Id
	//}).([]*modelpb.Model)
	//
	//// get model providers
	//var providerIds []string
	//for _, model := range allModels {
	//	providerIds = append(providerIds, model.ProviderId)
	//}
	//providerIds = strutil.DedupSlice(providerIds, true)
	//providerPagingResp, err := h.DAO.ModelProviderClient().Paging(ctx, &modelproviderpb.ModelProviderPagingRequest{
	//	PageNum:  1,
	//	PageSize: 999,
	//	Ids:      providerIds,
	//})
	//if err != nil {
	//	return nil, err
	//}
	//providerMapById := make(map[string]*modelproviderpb.ModelProvider)
	//for _, provider := range providerPagingResp.List {
	//	providerMapById[provider.Id] = provider
	//}

	allModels, err := cachehelpers.ListAllClientModels(ctx, client.Id)
	if err != nil {
		return nil, err
	}

	// assign rich models
	richModels := make([]*pb.RichModel, 0, len(allModels))
	metadataEnhancerService := i18n_services.NewMetadataEnhancerService(ctx, h.DAO)

	// get from ctx
	inputLang := string(i18n.LocaleDefault)
	if lang, ok := ctxhelper.GetAccessLang(ctx); ok {
		inputLang = lang
	}
	locale := i18n_services.GetLocaleFromContext(inputLang)
	for _, model := range allModels {
		// enhance model
		enhancedModel := metadataEnhancerService.EnhanceModelMetadata(ctx, model.Model, locale)
		richModels = append(richModels, &pb.RichModel{
			Model:    desensitiveModel(enhancedModel),
			Provider: desensitiveProvider(model.Provider),
		})
	}
	// assign rich client
	richClient := &pb.RichClient{
		Client: desensitiveClient(client),
		Models: richModels,
	}

	return richClient, nil
}
