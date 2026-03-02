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

package health

import (
	"net/http"
	"strings"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type APIType string

const (
	APITypeChatCompletions APIType = "chat_completions"
	APITypeResponses       APIType = "responses"
)

const (
	modelHealthBindingKey state_store.BindingKey = "global:model-health"
	stateUnhealthy                               = "unhealthy"
	stateHealthy                                 = "healthy"
)

type ModelHealthState struct {
	State     string    `json:"state"`
	APIType   APIType   `json:"api_type"`
	LastError string    `json:"last_error,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ResolveAPIType(method, path string) (APIType, bool) {
	if !strings.EqualFold(method, http.MethodPost) {
		return "", false
	}
	normalizedPath := strings.TrimSuffix(path, "/")
	switch normalizedPath {
	case vars.RequestPathPrefixV1ChatCompletions:
		return APITypeChatCompletions, true
	case vars.RequestPathPrefixV1Responses:
		return APITypeResponses, true
	default:
		return "", false
	}
}
