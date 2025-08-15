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

// OpenAIResponsesThinkingExtractor handles OpenAI Responses style thinking fields
// Fields: reasoning.effort
type OpenAIResponsesThinkingExtractor struct{}

func (e *OpenAIResponsesThinkingExtractor) ExtractMode(body map[string]any) (*types.CommonThinkingMode, error) {
	return nil, nil
}

func (e *OpenAIResponsesThinkingExtractor) ExtractEffort(body map[string]any) (*types.CommonThinkingEffort, error) {
	if reasoningObj, ok := body[types.FieldReasoning]; ok {
		if reasoningMap, ok := reasoningObj.(map[string]any); ok {
			if effortVal, ok := reasoningMap[types.FieldEffort]; ok {
				if effortStr, ok := effortVal.(string); ok {
					if types.IsValidEffort(effortStr) {
						return types.EffortPtr(types.CommonThinkingEffort(effortStr)), nil
					}
				}
			}
		}
	}
	return nil, nil
}

func (e *OpenAIResponsesThinkingExtractor) ExtractBudgetTokens(body map[string]any) (*int, error) {
	// openai responses doesn't have budget field
	return nil, nil
}

func (e *OpenAIResponsesThinkingExtractor) GetPriority() int {
	return 3
}

func (e *OpenAIResponsesThinkingExtractor) GetName() string {
	return "openai_responses"
}

func (e *OpenAIResponsesThinkingExtractor) CanExtract(body map[string]any) bool {
	_, hasReasoning := body[types.FieldReasoning]
	return hasReasoning
}

func (e *OpenAIResponsesThinkingExtractor) RelatedFields(body map[string]any) map[string]any {
	result := make(map[string]any)
	// extract reasoning object if present
	if reasoningObj, ok := body[types.FieldReasoning]; ok {
		result[types.FieldReasoning] = reasoningObj
	}
	return result
}
