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

package context

import (
	"context"
	"fmt"
	"net/http"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type RequestForModel struct {
	Model string `json:"model"`
}

// ModelWithProvider combines model with its provider information
type ModelWithProvider struct {
	*modelpb.Model
	Provider *providerpb.ModelProvider
}

func findModel(req *http.Request, requestCtx interface{}, client *clientpb.Client) (*modelpb.Model, error) {
	q := ctxhelper.MustGetDBClient(req.Context())

	// Use unified lookup function
	identifier, err := findModelIdentifier(req, requestCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to find model: %v", err)
	}
	if identifier == nil {
		return nil, fmt.Errorf("missing model")
	}

	// If there's a specific UUID, get model directly
	if identifier.ID != "" {
		model, err := q.ModelClient().Get(req.Context(), &modelpb.ModelGetRequest{Id: identifier.ID})
		if err != nil {
			return nil, fmt.Errorf("failed to get model by uuid: %s", identifier.ID)
		}
		return model, nil
	}

	// Find by model name + optional publisher
	if identifier.Name == "" {
		return nil, fmt.Errorf("missing model")
	}

	requiredModelPublisher := req.Header.Get(vars.XAIProxyModelPublisher)
	oneModel, err := getOneModelByNameAndPublisher(req.Context(), identifier.Name, requiredModelPublisher, client.Id)
	if err != nil {
		return nil, err
	}
	if oneModel != nil {
		return oneModel, nil
	}
	// model not found, construct friendly error
	return nil, constructFriendlyError(req.Context(), identifier.Name, requiredModelPublisher)
}

func getOneModelByNameAndPublisher(ctx context.Context, inputModelName string, inputModelPublisher string, clientID string) (*modelpb.Model, error) {
	allClientModels, err := listAllClientModels(ctx, clientID)
	if err != nil {
		return nil, err
	}
	modelsByKey := getMapOfAvailableNameWithModels(allClientModels)
	var inputKeysInOrder []string
	if inputModelPublisher != "" {
		inputKeysInOrder = append(inputKeysInOrder, fmt.Sprintf("%s/%s", inputModelPublisher, inputModelName))
	} else {
		inputKeysInOrder = append(inputKeysInOrder, inputModelName)
	}
	for _, inputKey := range inputKeysInOrder {
		foundModel := modelGetter(ctx, modelsByKey[inputKey])
		if foundModel != nil {
			return foundModel, nil
		}
	}
	return nil, nil
}

// modelGetter get one model, always return the last-updated model.
// TODO: Implement intelligent model selection based on:
// - Load balancing: distribute requests across available models
// - Cost optimization: prefer cheaper models when performance is comparable
// - Performance metrics: select models with better latency/throughput
// - Health status: avoid models with high error rates
func modelGetter(ctx context.Context, models []*modelpb.Model) *modelpb.Model {
	if len(models) == 0 {
		return nil
	}
	// sort models by updated_at desc
	var latestModel *modelpb.Model
	for _, model := range models {
		if latestModel == nil || model.UpdatedAt.AsTime().After(latestModel.UpdatedAt.AsTime()) {
			latestModel = model
		}
	}
	return latestModel
}

// getMapOfAvailableNameWithModels.
// availableName rules: (priority: high -> low)
// - ${publisher}/${model.name}
// - ${model.name}
// - ${model.provider.type}/${model.name}
func getMapOfAvailableNameWithModels(clientModels []*ModelWithProvider) map[string][]*modelpb.Model {
	modelsMap := make(map[string][]*modelpb.Model)
	for _, model := range clientModels {
		publisher := model.Metadata.Public["publisher"].GetStringValue()
		keys := append([]string{},
			fmt.Sprintf("%s/%s", publisher, model.Name),
			model.Name,
			fmt.Sprintf("%s/%s", model.Provider.Type, model.Name), // compatible with old format: model.Name + provider.Type
		)
		for _, key := range keys {
			if key == "" || key == "/" {
				continue
			}
			if _, ok := modelsMap[key]; !ok {
				modelsMap[key] = []*modelpb.Model{}
			}
			modelsMap[key] = append(modelsMap[key], model.Model)
		}
	}
	return modelsMap
}

func listAllModels(ctx context.Context) ([]*ModelWithProvider, error) {
	cacheManager := ctxhelper.MustGetCacheManager(ctx)

	// get models from cache
	modelsV, err := cacheManager.ListAll(ctx, cachetypes.ItemTypeModel)
	if err != nil {
		return nil, err
	}
	models := modelsV.([]*modelpb.Model)

	// get providers from cache
	providersV, err := cacheManager.ListAll(ctx, cachetypes.ItemTypeProvider)
	if err != nil {
		return nil, err
	}
	providers := providersV.([]*providerpb.ModelProvider)

	// build provider mapping for fast lookup
	providerMap := make(map[string]*providerpb.ModelProvider)
	for _, provider := range providers {
		providerMap[provider.Id] = provider
	}

	// combine models with providers
	var result []*ModelWithProvider
	for _, model := range models {
		provider, ok := providerMap[model.ProviderId]
		if !ok {
			continue // skip models without provider
		}
		result = append(result, &ModelWithProvider{
			Model:    model,
			Provider: provider,
		})
	}

	return result, nil
}

func listAllClientModels(ctx context.Context, clientID string) ([]*ModelWithProvider, error) {
	// get client model IDs directly from database (since it's per-client and changes frequently)
	q := ctxhelper.MustGetDBClient(ctx)
	clientModelsResp, err := q.ClientModelRelationClient().ListClientModels(ctx, &clientmodelrelationpb.ListClientModelsRequest{ClientId: clientID})
	if err != nil {
		return nil, fmt.Errorf("failed to list client models: %v", err)
	}

	if len(clientModelsResp.ModelIds) == 0 {
		return nil, nil
	}

	// get all models with providers
	allModels, err := listAllModels(ctx)
	if err != nil {
		return nil, err
	}

	// filter models by client relations
	var clientModels []*ModelWithProvider
	for _, model := range allModels {
		for _, modelID := range clientModelsResp.ModelIds {
			if model.Id == modelID {
				clientModels = append(clientModels, model)
				break
			}
		}
	}

	return clientModels, nil
}

func constructFriendlyError(ctx context.Context, inputModelName string, inputModelPublisher string) error {
	allModels, err := listAllModels(ctx)
	if err != nil {
		return err
	}
	modelsByKey := getMapOfAvailableNameWithModels(allModels)
	var inputKeysInOrder []string
	if inputModelPublisher != "" {
		inputKeysInOrder = append(inputKeysInOrder, fmt.Sprintf("%s/%s", inputModelPublisher, inputModelName))
	} else {
		inputKeysInOrder = append(inputKeysInOrder, inputModelName)
	}
	foundModelWithoutPermission := false
	for _, inputKey := range inputKeysInOrder {
		_, ok := modelsByKey[inputKey]
		if ok {
			foundModelWithoutPermission = true
			break
		}
	}
	var errMsg string
	if !foundModelWithoutPermission {
		errMsg = fmt.Sprintf("model not available: %s", inputModelName)
	} else {
		errMsg = fmt.Sprintf("you don't have permission to access the model: %s", inputModelName)
	}
	if inputModelPublisher != "" {
		errMsg += fmt.Sprintf(", publisher: %s", inputModelPublisher)
	}
	if !foundModelWithoutPermission {
		errMsg += ". If this is a newly added model, please wait a moment and try again."
	}
	return fmt.Errorf("%s", errMsg)
}
