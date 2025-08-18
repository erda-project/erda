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

// AnthropicThinkingExtractor handles Anthropic style thinking fields
// Fields: thinking.type, thinking.budget_tokens
type AnthropicThinkingExtractor struct{}

func (e *AnthropicThinkingExtractor) ExtractMode(body map[string]any) (*types.CommonThinkingMode, error) {
	if thinkingObj, ok := body[types.FieldThinking]; ok {
		if thinkingMap, ok := thinkingObj.(map[string]any); ok {
			if typeVal, ok := thinkingMap[types.FieldType]; ok {
				if typeStr, ok := typeVal.(string); ok {
					switch typeStr {
					case "enabled":
						return types.ModePtr(types.ModeOn), nil
					case "disabled":
						return types.ModePtr(types.ModeOff), nil
					case "auto":
						return types.ModePtr(types.ModeAuto), nil
					}
				}
			}
		}
	}
	return nil, nil
}

func (e *AnthropicThinkingExtractor) ExtractEffort(body map[string]any) (*types.CommonThinkingEffort, error) {
	// anthropic doesn't have effort field
	return nil, nil
}

func (e *AnthropicThinkingExtractor) ExtractBudgetTokens(body map[string]any) (*int, error) {
	if thinkingObj, ok := body[types.FieldThinking]; ok {
		if thinkingMap, ok := thinkingObj.(map[string]any); ok {
			if budgetVal, ok := thinkingMap[types.FieldBudgetTokens]; ok {
				return extractors.ExtractIntValue(budgetVal), nil
			}
		}
	}
	return nil, nil
}

func (e *AnthropicThinkingExtractor) GetPriority() int {
	return 1 // highest priority
}

func (e *AnthropicThinkingExtractor) GetName() string {
	return "anthropic"
}

func (e *AnthropicThinkingExtractor) CanExtract(body map[string]any) bool {
	_, hasThinking := body[types.FieldThinking]
	return hasThinking
}

func (e *AnthropicThinkingExtractor) RelatedFields(body map[string]any) map[string]any {
	result := make(map[string]any)
	// extract thinking object if present
	if thinkingObj, ok := body[types.FieldThinking]; ok {
		result[types.FieldThinking] = thinkingObj
	}
	return result
}
