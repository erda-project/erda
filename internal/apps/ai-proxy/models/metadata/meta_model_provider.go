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

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_style"
)

type (
	ModelProviderMeta struct {
		Public ModelProviderMetaPublic `json:"public,omitempty"`
		Secret ModelProviderMetaSecret `json:"secret,omitempty"`
	}
	ModelProviderMetaPublic struct {
		Scheme   string `json:"scheme,omitempty"`   // deprecated, use @APIStyleConfig
		Endpoint string `json:"endpoint,omitempty"` // deprecated, use @APIStyleConfig
		Host     string `json:"host,omitempty"`     // deprecated, use @APIStyleConfig
		Location string `json:"location,omitempty"`

		RewritePath string `json:"rewritePath,omitempty"`

		// API Related configs
		API *API `json:"api,omitempty"`
	}
	ModelProviderMetaSecret struct {
		AnotherAPIKey string `json:"anotherApiKey,omitempty"`
	}
)

type API struct {
	APIStyle       api_style.APIStyle        `json:"apiStyle,omitempty"`
	APIStyleConfig *api_style.APIStyleConfig `json:"apiStyleConfig,omitempty"`
}

func (m *Metadata) ToModelProviderMeta() (*ModelProviderMeta, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Metadata to json: %v", err)
	}
	var result ModelProviderMeta
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal string to ModelProviderMeta: %v", err)
	}
	return &result, nil
}

func (m *Metadata) MustToModelProviderMeta() *ModelProviderMeta {
	meta, err := m.ToModelProviderMeta()
	if err != nil {
		panic(fmt.Sprintf("failed to convert Metadata to ModelProviderMeta: %v", err))
	}
	return meta
}
