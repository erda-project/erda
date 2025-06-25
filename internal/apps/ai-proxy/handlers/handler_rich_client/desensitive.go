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
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
)

// desensitiveModel removes sensitive information from model in-place
func desensitiveModel(model *modelpb.Model) *modelpb.Model {
	if model == nil {
		return nil
	}
	model.ApiKey = ""
	if model.Metadata != nil {
		model.Metadata.Secret = nil
	}
	return model
}

// desensitiveProvider removes sensitive information from provider in-place
func desensitiveProvider(provider *modelproviderpb.ModelProvider) *modelproviderpb.ModelProvider {
	if provider == nil {
		return nil
	}
	provider.ApiKey = ""
	if provider.Metadata != nil {
		provider.Metadata.Secret = nil
		// hide api configs
		if provider.Metadata.Public != nil {
			delete(provider.Metadata.Public, "api")
		}
	}
	return provider
}

func desensitiveClient(client *clientpb.Client) *clientpb.Client {
	if client == nil {
		return nil
	}
	client.AccessKeyId = ""
	client.SecretKeyId = ""
	if client.Metadata != nil {
		client.Metadata.Secret = nil
	}
	return client
}
