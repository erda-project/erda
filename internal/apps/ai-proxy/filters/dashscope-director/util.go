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

package dashscope_director

import (
	"fmt"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

func getProviderMeta(p *modelproviderpb.ModelProvider) metadata.ModelProviderMeta {
	var meta metadata.ModelProviderMeta
	cputil.MustObjJSONTransfer(&p.Metadata, &meta)
	return meta
}

func getModelMeta(m *modelpb.Model) metadata.AliyunDashScopeModelMeta {
	var meta metadata.AliyunDashScopeModelMeta
	cputil.MustObjJSONTransfer(&m.Metadata, &meta)
	return meta
}

func mustGetResponseType(modelMeta metadata.AliyunDashScopeModelMeta) metadata.AliyunDashScopeResponseType {
	requestType := modelMeta.Public.RequestType
	responseType := modelMeta.Public.ResponseType
	if responseType == "" {
		responseType = metadata.AliyunDashScopeResponseType(requestType.String())
	}

	if ok, err := responseType.Valid(); !ok {
		panic(fmt.Errorf("metadata.public.response_type is invalid, err: %v", err))
	}

	return responseType
}
