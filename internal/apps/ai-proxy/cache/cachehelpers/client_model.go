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

package cachehelpers

import (
	"context"
	"fmt"

	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// ModelWithProvider combines model with its provider information
type ModelWithProvider struct {
	*modelpb.Model
	Provider *providerpb.ServiceProvider
}

func GetOneClientModel(ctx context.Context, clientID, modelID string) (*ModelWithProvider, error) {
	allClientModels, err := ListAllClientModels(ctx, clientID)
	if err != nil {
		return nil, err
	}
	for _, clientModel := range allClientModels {
		if clientModel.Id == modelID {
			return clientModel, nil
		}
	}
	return nil, fmt.Errorf("failed to get model by id: %s", modelID)
}

func ListAllClientModels(ctx context.Context, clientID string) ([]*ModelWithProvider, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)

	// platform-assigned models
	assignedModels, err := _listAllClientAssignedModels(ctx, clientID)
	if err != nil {
		return nil, err
	}

	// client-belonged models
	belongedModels, err := _listAllClientBelongedModels(ctx, clientID)
	if err != nil {
		return nil, err
	}

	// all
	allModels := append(assignedModels, belongedModels...)
	providerIDMap := make(map[string]struct{})
	providerMap := make(map[string]*providerpb.ServiceProvider)
	for _, model := range allModels {
		_, ok := providerMap[model.ProviderId]
		if ok {
			continue
		}
		providerIDMap[model.ProviderId] = struct{}{}
		providerV, err := cache.GetByID(ctx, cachetypes.ItemTypeProvider, model.ProviderId)
		if err != nil {
			return nil, err
		}
		providerMap[model.ProviderId] = providerV.(*providerpb.ServiceProvider)
	}
	var allModelsWithProvider []*ModelWithProvider
	for _, model := range allModels {
		allModelsWithProvider = append(allModelsWithProvider, &ModelWithProvider{
			Model:    model,
			Provider: providerMap[model.ProviderId],
		})
	}

	return allModelsWithProvider, nil
}

func _listAllClientBelongedModels(ctx context.Context, clientID string) ([]*modelpb.Model, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	_, allModelsV, err := cache.ListAll(ctx, cachetypes.ItemTypeModel)
	if err != nil {
		return nil, err
	}
	allModels := allModelsV.([]*modelpb.Model)
	allClientBelongsModels := make([]*modelpb.Model, 0, len(allModels))
	for _, model := range allModels {
		if model.ClientId == clientID {
			allClientBelongsModels = append(allClientBelongsModels, model)
		}
	}
	return allClientBelongsModels, nil
}

func _listAllClientAssignedModels(ctx context.Context, clientID string) ([]*modelpb.Model, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	_, relationsV, err := cache.ListAll(ctx, cachetypes.ItemTypeClientModelRelation)
	if err != nil {
		return nil, fmt.Errorf("failed to list client model relations: %v", err)
	}
	var matchedRelations []*clientmodelrelationpb.ClientModelRelation
	for _, relation := range relationsV.([]*clientmodelrelationpb.ClientModelRelation) {
		if relation.ClientId == clientID {
			matchedRelations = append(matchedRelations, relation)
		}
	}
	if len(matchedRelations) == 0 {
		return nil, nil
	}
	assignedModels := make([]*modelpb.Model, 0, len(matchedRelations))
	for _, relation := range matchedRelations {
		modelV, err := cache.GetByID(ctx, cachetypes.ItemTypeModel, relation.ModelId)
		if err != nil {
			return nil, err
		}
		model := modelV.(*modelpb.Model)
		assignedModels = append(assignedModels, model)
	}

	return assignedModels, nil
}
