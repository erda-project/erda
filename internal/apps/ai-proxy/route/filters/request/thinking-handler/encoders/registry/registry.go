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

package registry

import (
	"context"
	"fmt"
	"sort"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders"
	mp "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders/impl/model-publisher"
	sp "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders/impl/service-provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// Registry manages all CommonThinking encoders
type Registry struct {
	encoders []encoders.CommonThinkingEncoder
}

// NewRegistry creates a new registry with default encoders
func NewRegistry() *Registry {
	encoderList := []encoders.CommonThinkingEncoder{
		// service provider
		&sp.VolcengineThinkingEncoder{},
		&sp.BailianThinkingEncoder{},
		// model publisher
		&mp.AnthropicThinkingEncoder{},
		&mp.QwenThinkingEncoder{},
		&mp.OpenAIChatThinkingEncoder{},
		&mp.OpenAIResponsesThinkingEncoder{},
		&mp.BytedanceThinkingEncoder{},
	}

	// sort by priority (lower number = higher priority)
	sort.Slice(encoderList, func(i, j int) bool {
		return encoderList[i].GetPriority() < encoderList[j].GetPriority()
	})

	return &Registry{
		encoders: encoderList,
	}
}

// GetEncoders returns all encoders sorted by priority
func (r *Registry) GetEncoders() []encoders.CommonThinkingEncoder {
	return r.encoders
}

// EncodeAll attempts to encode using the first applicable encoder
// Returns error if thinking configuration is needed but no encoder can handle it
func (r *Registry) EncodeAll(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	// if no thinking configuration provided, nothing to encode
	if ct.Mode == nil && ct.Effort == nil && ct.BudgetTokens == nil {
		return nil, nil
	}
	for _, encoder := range r.encoders {
		if encoder.CanEncode(ctx) {
			return encoder.Encode(ctx, ct)
		}
	}
	return nil, fmt.Errorf("no thinking encoder found")
}
