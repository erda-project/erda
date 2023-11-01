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

package bailian_director

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

var clientsMapByProviderID map[string]*CompletionClient

func init() {
	clientsMapByProviderID = make(map[string]*CompletionClient)
}

func getProviderMeta(p *modelproviderpb.ModelProvider) metadata.AliyunBailianProviderMeta {
	var meta metadata.AliyunBailianProviderMeta
	cputil.MustObjJSONTransfer(&p.Metadata, &meta)
	return meta
}

func getModelMeta(m *modelpb.Model) metadata.AliyunBailianModelMeta {
	var meta metadata.AliyunBailianModelMeta
	cputil.MustObjJSONTransfer(&m.Metadata, &meta)
	return meta
}

func fetchClient(p *modelproviderpb.ModelProvider) *CompletionClient {
	if c, ok := clientsMapByProviderID[p.Id]; ok {
		return c
	}
	meta := getProviderMeta(p)
	client := &CompletionClient{
		AccessKeyId:     &meta.Secret.AccessKeyId,
		AccessKeySecret: &meta.Secret.AccessKeySecret,
		AgentKey:        &meta.Secret.AgentKey,
	}
	clientsMapByProviderID[p.Id] = client
	return client
}
