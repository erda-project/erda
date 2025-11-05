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
		_client, err := cachehelpers.GetClientByAK(ctx, req.AccessKeyId)
		if err != nil {
			return nil, err
		}
		client = _client
	}

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
		// render template
		renderedModel, err := cachehelpers.GetRenderedModelByID(ctx, model.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to render model, id: %s, err: %w", model.Id, err)
		}
		// enhance model
		enhancedModel := metadataEnhancerService.EnhanceModelMetadata(ctx, renderedModel, locale)
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
