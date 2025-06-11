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
	clientMatchedModels, err := getClientModelsByNameAndPublisher(ctx, client.Id, reqForModelName.Model, requiredModelPublisher)
	if err != nil {
		return nil, fmt.Errorf("failed to get client models by name and publisher: %v", err)
	}
	// if there are multiple, take the first one, default is sorted by updated_at desc
	if len(clientMatchedModels) > 0 {
		return clientMatchedModels[0], nil
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

func getClientModelsByNameAndPublisher(ctx context.Context, clientID string, modelName string, publisher string) ([]*modelpb.Model, error) {
	q := ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
	clientModelsResp, err := q.ClientModelRelationClient().ListClientModels(ctx, &clientmodelrelationpb.ListClientModelsRequest{
		ClientId: clientID,
	})
	if err != nil {
		return nil, err
	}
	if len(clientModelsResp.ModelIds) == 0 {
		return nil, nil
	}
	// list models by name & ids
	pagingModelResp, err := q.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
		PageNum:  1,
		PageSize: 999,
		NameFull: modelName,
		Ids:      clientModelsResp.ModelIds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to paging models by name and modelIDs, name: %s, err: %v", modelName, err)
	}
	// filter by publisher
	return filterModelsByPublisher(pagingModelResp.List, publisher), nil
}

func filterModelsByPublisher(in []*modelpb.Model, publisher string) (out []*modelpb.Model) {
	if publisher == "" {
		return in
	}
	for _, model := range in {
		modelPublisher := model.Metadata.Public["publisher"].GetStringValue()
		if modelPublisher == publisher {
			out = append(out, model)
		}
	}
	return
}

func getModelsByNameAndPublisher(ctx context.Context, modelName string, requiredModelPublisher string) ([]*modelpb.Model, error) {
	q := ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
	sameNameModelsResp, err := q.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
		PageNum:  1,
		PageSize: 999,
		NameFull: modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get models by model name: %s, err: %v", modelName, err)
	}
	// filter by publisher
	return filterModelsByPublisher(sameNameModelsResp.List, requiredModelPublisher), nil
}

func constructFriendlyError(ctx context.Context, modelName string, requiredModelPublisher string) error {
	models, _ := getModelsByNameAndPublisher(ctx, modelName, requiredModelPublisher)
	var errMsg string
	if len(models) == 0 {
		errMsg = fmt.Sprintf("model not exist: %s", modelName)
	} else {
		errMsg = fmt.Sprintf("you don't have permission to access the model: %s", modelName)
	}
	if requiredModelPublisher != "" {
		errMsg += fmt.Sprintf(", publisher: %s", requiredModelPublisher)
	}
	return fmt.Errorf(errMsg)
}
