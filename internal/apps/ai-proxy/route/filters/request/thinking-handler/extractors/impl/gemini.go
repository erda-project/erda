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

// GeminiThinkingExtractor handles gemini style thinking fields
// Fields: thinking_config
type GeminiThinkingExtractor struct{}

func (e *GeminiThinkingExtractor) ExtractMode(body map[string]any) (*types.CommonThinkingMode, error) {
	return nil, nil
}

func (e *GeminiThinkingExtractor) ExtractEffort(body map[string]any) (*types.CommonThinkingEffort, error) {
	// gemini doesn't have effort field
	return nil, nil
}

func (e *GeminiThinkingExtractor) ExtractBudgetTokens(body map[string]any) (*int, error) {
	if thinkingObj, ok := body[types.FieldThinkingConfig]; ok {
		if thinkingMap, ok := thinkingObj.(map[string]any); ok {
			if budgetVal, ok := thinkingMap["thinking_budget"]; ok {
				return extractors.ExtractIntValue(budgetVal), nil
			}
		}
	}
	return nil, nil
}

func (e *GeminiThinkingExtractor) GetPriority() int {
	return 5
}

func (e *GeminiThinkingExtractor) GetName() string {
	return "gemini"
}

func (e *GeminiThinkingExtractor) CanExtract(body map[string]any) bool {
	_, hasReasoningEffort := body[types.FieldThinkingConfig]
	return hasReasoningEffort
}

func (e *GeminiThinkingExtractor) RelatedFields(body map[string]any) map[string]any {
	result := make(map[string]any)
	// extract thinking_config if present
	if thinkingConfig, ok := body[types.FieldThinkingConfig]; ok {
		result[types.FieldThinkingConfig] = thinkingConfig
	}
	return result
}
