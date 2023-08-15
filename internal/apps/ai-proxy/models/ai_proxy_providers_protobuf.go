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

package models

import (
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
)

func (a *AIProxyProviders) FromProtobuf(provider *pb.Provider) *AIProxyProviders {
	*a = AIProxyProviders{
		Name:        provider.GetName(),
		InstanceID:  provider.GetInstanceId(),
		Host:        provider.GetHost(),
		Scheme:      provider.GetScheme(),
		Description: provider.GetDescription(),
		DocSite:     provider.GetDocSite(),
		AesKey:      "",
		APIKey:      "",
		Metadata:    provider.GetMetadata(),
	}
	a.ResetAesKey()
	a.SetAPIKeyWithEncrypt(provider.GetApiKey())
	return a
}

func (a *AIProxyProviders) ToProtobuf() *pb.Provider {
	return &pb.Provider{
		Name:        a.Name,
		InstanceId:  a.Host,
		Host:        a.Host,
		Scheme:      a.Scheme,
		Description: a.Description,
		DocSite:     a.DocSite,
		ApiKey:      a.GetAPIKeyWithDecrypt(),
		Metadata:    a.Metadata,
	}
}

func (list AIProxyProvidersList) ToProtobuf() []*pb.Provider {
	var length = len(list)
	var results = make([]*pb.Provider, length)
	for i := 0; i < length; i++ {
		results[i] = list[i].ToProtobuf()
	}
	return results
}
