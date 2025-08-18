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

package impl

import (
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/extractors"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// QwenThinkingExtractor handles Qwen style thinking fields
// Fields: enable_thinking, thinking_budget
type QwenThinkingExtractor struct{}

func (e *QwenThinkingExtractor) ExtractMode(body map[string]any) (*types.CommonThinkingMode, error) {
	if enableThinking, ok := body[types.FieldEnableThinking]; ok {
		if enabled, ok := enableThinking.(bool); ok {
			if enabled {
				return types.ModePtr(types.ModeOn), nil
			} else {
				return types.ModePtr(types.ModeOff), nil
			}
		}
	}
	return nil, nil
}

func (e *QwenThinkingExtractor) ExtractEffort(body map[string]any) (*types.CommonThinkingEffort, error) {
	// qwen doesn't have effort field
	return nil, nil
}

func (e *QwenThinkingExtractor) ExtractBudgetTokens(body map[string]any) (*int, error) {
	if thinkingBudget, ok := body[types.FieldThinkingBudget]; ok {
		return extractors.ExtractIntValue(thinkingBudget), nil
	}
	return nil, nil
}

func (e *QwenThinkingExtractor) GetPriority() int {
	return 2
}

func (e *QwenThinkingExtractor) GetName() string {
	return "qwen"
}

func (e *QwenThinkingExtractor) CanExtract(body map[string]any) bool {
	_, hasEnableThinking := body[types.FieldEnableThinking]
	_, hasThinkingBudget := body[types.FieldThinkingBudget]
	return hasEnableThinking || hasThinkingBudget
}

func (e *QwenThinkingExtractor) RelatedFields(body map[string]any) map[string]any {
	result := make(map[string]any)
	// extract enable_thinking if present
	if enableThinking, ok := body[types.FieldEnableThinking]; ok {
		result[types.FieldEnableThinking] = enableThinking
	}
	// extract thinking_budget if present
	if thinkingBudget, ok := body[types.FieldThinkingBudget]; ok {
		result[types.FieldThinkingBudget] = thinkingBudget
	}
	return result
}
