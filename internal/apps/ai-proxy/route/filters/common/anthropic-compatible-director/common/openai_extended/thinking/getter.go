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

package thinking

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/common/openai_extended"
)

// UnifiedGetThinkingConfigs get thinking configs from extra fields.
// Currently, we support below styles:
// - Anthropic Thinking
// - Qwen
func UnifiedGetThinkingConfigs(req openai_extended.OpenAIRequestExtended) *UnifiedThinking {
	if len(req.ExtraFields) == 0 {
		return nil
	}

	// get first non-nil thinking configs
	getters := []ThinkingGetter{
		getThinkingFromAnthropicStyle,
		getThinkingFromQwenStyle,
	}
	for _, getter := range getters {
		thinking := getter(req)
		if thinking != nil {
			return thinking
		}
	}
	return nil
}

type ThinkingGetter func(req openai_extended.OpenAIRequestExtended) *UnifiedThinking

var getThinkingFromAnthropicStyle = func(req openai_extended.OpenAIRequestExtended) *UnifiedThinking {
	var anthropicThinking AnthropicThinking
	cputil.MustObjJSONTransfer(req.ExtraFields, &anthropicThinking)
	if anthropicThinking.Thinking == nil {
		return nil
	}
	return &UnifiedThinking{Thinking: anthropicThinking.Thinking}
}

var getThinkingFromQwenStyle = func(req openai_extended.OpenAIRequestExtended) *UnifiedThinking {
	var qwenThinking QwenThinking
	cputil.MustObjJSONTransfer(req.ExtraFields, &qwenThinking)
	if qwenThinking.EnableThinking == nil {
		return nil
	}
	return &UnifiedThinking{Thinking: &AnthropicThinkingInternal{
		Type: func() string {
			if *qwenThinking.EnableThinking {
				return "enabled"
			}
			return "disabled"
		}(),
		BudgetTokens: qwenThinking.ThinkingBudget,
	}}
}
