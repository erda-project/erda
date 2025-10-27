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

package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment"
)

type (
	ServiceProviderMeta struct {
		Public ServiceProviderMetaPublic `json:"public,omitempty"`
		Secret ServiceProviderMetaSecret `json:"secret,omitempty"`
	}
	ServiceProviderMetaPublic struct {
		Scheme   string `json:"scheme,omitempty"`   // deprecated, use @APIStyleConfig
		Endpoint string `json:"endpoint,omitempty"` // deprecated, use @APIStyleConfig
		Host     string `json:"host,omitempty"`     // deprecated, use @APIStyleConfig
		Location string `json:"location,omitempty"`

		RewritePath string `json:"rewritePath,omitempty"`

		// API Related configs
		API *api_segment.API `json:"api,omitempty"`
	}
	ServiceProviderMetaSecret struct {
		AnotherAPIKey string `json:"anotherApiKey,omitempty"`
	}
)

func (m *Metadata) ToServiceProviderMeta() (*ServiceProviderMeta, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Metadata to json: %v", err)
	}
	var result ServiceProviderMeta
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal string to ServiceProviderMeta: %v", err)
	}
	return &result, nil
}

func (m *Metadata) MustToServiceProviderMeta() *ServiceProviderMeta {
	meta, err := m.ToServiceProviderMeta()
	if err != nil {
		panic(fmt.Sprintf("failed to convert Metadata to ServiceProviderMeta: %v", err))
	}
	return meta
}
