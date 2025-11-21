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
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// OpenAIChatThinkingExtractor handles OpenAI Chat style thinking fields
// Fields: reasoning_effort
type OpenAIChatThinkingExtractor struct{}

func (e *OpenAIChatThinkingExtractor) ExtractMode(body map[string]any) (*types.CommonThinkingMode, error) {
	if reasoningEffort, ok := body[types.FieldReasoningEffort]; ok {
		if effortStr, ok := reasoningEffort.(string); ok {
			if types.IsValidEffort(effortStr) {
				switch effortStr {
				case types.EffortNone.String():
					return types.ModePtr(types.ModeOff), nil
				default:
					return types.ModePtr(types.ModeOn), nil
				}
			}
		}
	}
	return nil, nil
}

func (e *OpenAIChatThinkingExtractor) ExtractEffort(body map[string]any) (*types.CommonThinkingEffort, error) {
	if reasoningEffort, ok := body[types.FieldReasoningEffort]; ok {
		if effortStr, ok := reasoningEffort.(string); ok {
			if types.IsValidEffort(effortStr) {
				return types.EffortPtr(types.CommonThinkingEffort(effortStr)), nil
			}
		}
	}
	return nil, nil
}

func (e *OpenAIChatThinkingExtractor) ExtractBudgetTokens(body map[string]any) (*int, error) {
	// openai chat doesn't have budget field
	return nil, nil
}

func (e *OpenAIChatThinkingExtractor) GetPriority() int {
	return 4
}

func (e *OpenAIChatThinkingExtractor) GetName() string {
	return "openai_chat"
}

func (e *OpenAIChatThinkingExtractor) CanExtract(body map[string]any) bool {
	_, hasReasoningEffort := body[types.FieldReasoningEffort]
	return hasReasoningEffort
}

func (e *OpenAIChatThinkingExtractor) RelatedFields(body map[string]any) map[string]any {
	result := make(map[string]any)
	// extract reasoning_effort if present
	if reasoningEffort, ok := body[types.FieldReasoningEffort]; ok {
		result[types.FieldReasoningEffort] = reasoningEffort
	}
	return result
}
