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

package openai_extended

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type OpenAIRequestExtended struct {
	openai.ChatCompletionRequest

	// ExtraFields is used to store extra fields that are not part of the original OpenAI request.
	ExtraFields map[string]any `json:"extra_fields,omitempty"` // used to store extra fields that are not part of the original OpenAI request
}

func (m *OpenAIRequestExtended) UnmarshalJSON(data []byte) error {
	type Alias OpenAIRequestExtended // avoid infinite recursion
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// decode to map, to extract extra fields
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// dynamically collect known fields from ChatCompletionRequest
	known := make(map[string]struct{})
	t := reflect.TypeOf(openai.ChatCompletionRequest{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag != "" {
			name := strings.Split(tag, ",")[0]
			if name != "" && name != "-" {
				known[name] = struct{}{}
			}
		}
	}

	m.ExtraFields = make(map[string]any)
	for k, v := range raw {
		if _, ok := known[k]; !ok {
			m.ExtraFields[k] = v
		}
	}

	return nil
}
