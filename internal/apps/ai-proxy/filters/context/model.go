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
	"encoding/json"
	"fmt"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	openai_v1_models "github.com/erda-project/erda/internal/apps/ai-proxy/filters/openai-v1-models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

type RequestForModel struct {
	Model string `json:"model"`
}

func findModel(ctx context.Context, infor reverseproxy.HttpInfor, client *clientpb.Client) (*modelpb.Model, error) {
	r := infor.Request()
	q := ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)

	// find model uuid, get exactly by model uuid;
	modelID, err := findModelID(infor)
	if err != nil {
		return nil, fmt.Errorf("failed to get model id, err: %v", err)
	}
	// if concrete uuid specified, no need to check client model relation
	if modelID != "" {
		model, err := q.ModelClient().Get(ctx, &modelpb.ModelGetRequest{Id: modelID})
		if err != nil {
			return nil, fmt.Errorf("failed to get model by uuid: %s", modelID)
		}
		return model, nil
	}

	// find by model name + optional publisher
	var reqForModelName RequestForModel
	bodyCopy := infor.BodyBuffer(true)
	if err := json.NewDecoder(bodyCopy).Decode(&reqForModelName); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}
	if reqForModelName.Model == "" {
		return nil, fmt.Errorf("missing required model field in request body")
	}
	requiredModelPublisher := r.Header.Get(vars.XAIProxyModelPublisher)
	oneModel, err := getOneModelByNameAndPublisher(ctx, reqForModelName.Model, requiredModelPublisher, client.Id)
	if err != nil {
		return nil, err
	}
	if oneModel != nil {
		return oneModel, nil
	}
	// model not found, construct friendly error
	return nil, constructFriendlyError(ctx, reqForModelName.Model, requiredModelPublisher)
}

func findModelID(infor reverseproxy.HttpInfor) (string, error) {
	r := infor.Request()
	// header preferred
	headerModelId := r.Header.Get(vars.XAIProxyModelId)
	if headerModelId != "" {
		return headerModelId, nil
	}
	// get from body
	// support json body
	if r.Header.Get(httputil.HeaderKeyContentType) == string(httputil.ApplicationJson) {
		snapshotBody := infor.BodyBuffer(true)
		var reqForModelID RequestForModel
		if err := json.NewDecoder(snapshotBody).Decode(&reqForModelID); err != nil {
			return "", fmt.Errorf("failed to decode request body, err: %v", err)
		}
		if reqForModelID.Model != "" {
			// parse truly model uuid, which is generated at api `/v1/models`/, see: internal/apps/ai-proxy/filters/openai-v1-models/filter.go#generateModelDisplayName
			uuid := openai_v1_models.ParseModelUUIDFromDisplayName(reqForModelID.Model)
			if uuid != "" {
				return uuid, nil
			}
		}
	}
	return "", nil
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
// - ${publisher}/${model.metadata.public.model_id}
// - ${publisher}/${model.metadata.public.model_name}
// - ${model.name}
func getMapOfAvailableNameWithModels(clientModels []*modelpb.Model) map[string][]*modelpb.Model {
	modelsMap := make(map[string][]*modelpb.Model)
	for _, model := range clientModels {
		publisher := model.Metadata.Public["publisher"].GetStringValue()
		modelIDInMetadata := model.Metadata.Public["model_id"].GetStringValue()
		modelNameInMetadata := model.Metadata.Public["model_name"].GetStringValue()
		keys := append([]string{},
			fmt.Sprintf("%s/%s", publisher, model.Name),
			fmt.Sprintf("%s/%s", publisher, modelIDInMetadata),
			fmt.Sprintf("%s/%s", publisher, modelNameInMetadata),
			model.Name,
		)
		for _, key := range keys {
			if key == "" || key == "/" {
				continue
			}
			if _, ok := modelsMap[key]; !ok {
				modelsMap[key] = []*modelpb.Model{}
			}
			modelsMap[key] = append(modelsMap[key], model)
		}
	}
	return modelsMap
}

func listAllModels(ctx context.Context) ([]*modelpb.Model, error) {
	q := ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
	pagingModelResp, err := q.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
		PageNum:  1,
		PageSize: 999,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to paging all models, err: %v", err)
	}
	return pagingModelResp.List, nil
}

func listAllClientModels(ctx context.Context, clientID string) ([]*modelpb.Model, error) {
	// find client model relations
	q := ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
	clientModelsResp, err := q.ClientModelRelationClient().ListClientModels(ctx, &clientmodelrelationpb.ListClientModelsRequest{ClientId: clientID})
	if err != nil {
		return nil, fmt.Errorf("failed to list client models, err: %v", err)
	}
	if len(clientModelsResp.ModelIds) == 0 {
		return nil, nil
	}
	// filter from all models
	allModels, err := listAllModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all models, err: %v", err)
	}
	var clientModels []*modelpb.Model
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
		errMsg = fmt.Sprintf("model not exist: %s", inputModelName)
	} else {
		errMsg = fmt.Sprintf("you don't have permission to access the model: %s", inputModelName)
	}
	if inputModelPublisher != "" {
		errMsg += fmt.Sprintf(", publisher: %s", inputModelPublisher)
	}
	return fmt.Errorf(errMsg)
}
