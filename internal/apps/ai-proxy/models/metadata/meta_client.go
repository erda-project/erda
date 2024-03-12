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

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
)

type (
	ClientMeta struct {
		Public ClientMetaPublic `json:"public,omitempty"`
		Secret ClientMetaSecret `json:"secret,omitempty"`
	}
	ClientMetaPublic struct {
		DefaultModelIdOfTextGeneration string `json:"default_model_id_of_text_generation,omitempty"`
		DefaultModelIdOfImage          string `json:"default_model_id_of_image,omitempty"`
		DefaultModelIdOfAudio          string `json:"default_model_id_of_audio,omitempty"`
		DefaultModelIdOfEmbedding      string `json:"default_model_id_of_embedding,omitempty"`
		DefaultModelIdOfTextModeration string `json:"default_model_id_of_text_moderation,omitempty"`
		DefaultModelIdOfMultimodal     string `json:"default_model_id_of_multimodal,omitempty"`
		DefaultModelIdOfAssistant      string `json:"default_model_id_of_assistant,omitempty"`
	}
	ClientMetaSecret struct {
	}
)

func (m *Metadata) ToClientMeta() (*ClientMeta, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Metadata to json: %v", err)
	}
	var result ClientMeta
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal string to ClientMeta: %v", err)
	}
	return &result, nil
}

func (p *ClientMetaPublic) GetDefaultModelIdByModelType(modelType modelpb.ModelType) (string, bool) {
	if p == nil {
		return "", false
	}
	b, _ := json.Marshal(p)
	m := make(map[string]string)
	_ = json.Unmarshal(b, &m)
	defaultModelId, ok := m["default_model_id_of_"+modelType.String()]
	return defaultModelId, ok
}
