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

package service_provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// GoogleVertexAIThinkingEncoder handles google-vertex-ai thinking encoding
// Fields: extra_body.google.thinking_config
// see:
// - https://ai.google.dev/gemini-api/docs/openai#thinking
//
// gemini-3 use thinking_level
// gemini-2.5 use thinking_budget
type GoogleVertexAIThinkingEncoder struct{}

func (e *GoogleVertexAIThinkingEncoder) CanEncode(ctx context.Context) bool {
	model := ctxhelper.MustGetModel(ctx)
	return strings.EqualFold(model.Publisher, string(common_types.ModelPublisherGoogle))
}

func (e *GoogleVertexAIThinkingEncoder) Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	googleConfig := map[string]any{}
	appendBody := map[string]any{
		types.FieldExtraBody: map[string]any{
			types.FieldGoogle: googleConfig,
		},
	}
	if ct.MustGetMode() == types.ModeAuto {
		return nil, nil
	}
	if ct.MustGetMode() == types.ModeOff {
		googleConfig[types.FieldThinkingConfig] = map[string]any{types.FieldIncludeThoughts: false}
		return appendBody, nil
	}
	// gemini-3 use thinking_level
	// gemini-2.5 use thinking_budget
	model := ctxhelper.MustGetModel(ctx)
	useThinkingLevel := false
	if strings.Contains(strings.ToLower(model.Name), "gemini-3") {
		useThinkingLevel = true
	}
	// calculate suitable budget or level
	suitableBudget := 1024
	suitableLevel := types.EffortLow
	// get from budget_tokens if provided
	if ct.BudgetTokens != nil {
		suitableBudget = *ct.BudgetTokens
		suitableLevel = types.MapBudgetTokensToEffort(*ct.BudgetTokens)
		suitableLevel = normalizeEffortLevelForGemini(suitableLevel)
		goto RESULT
	}
	// get from effort if provided
	if ct.Effort != nil {
		suitableBudget = types.MapEffortToBudgetTokens(*ct.Effort)
		suitableLevel = normalizeEffortLevelForGemini(*ct.Effort)
		goto RESULT
	}
	// no suitable budget found, use default 1024
RESULT:
	thinkingConfig := map[string]any{
		types.FieldIncludeThoughts: true,
	}
	if useThinkingLevel {
		thinkingConfig[types.FieldThinkingLevel] = suitableLevel
	} else {
		thinkingConfig[types.FieldThinkingBudget] = suitableBudget
	}
	googleConfig[types.FieldThinkingConfig] = thinkingConfig
	googleConfig[types.FieldThoughtTagMarker] = "think"

	return appendBody, nil
}

func (e *GoogleVertexAIThinkingEncoder) GetPriority() int {
	return 5
}

func (e *GoogleVertexAIThinkingEncoder) GetName() string {
	return fmt.Sprintf("service_provider: %s", common_types.ServiceProviderTypeGoogleVertexAI.String())
}

// normalizeEffortLevelForGemini
// reasoning_effort (OpenAI)	thinking_level (Gemini 3)	thinking_budget (Gemini 2.5)
//
//	minimal                         low                       1,024
//	low                             low                       1,024
//	medium                          high                      8,192
//	high                            high                      24,576
func normalizeEffortLevelForGemini(inputEffort types.CommonThinkingEffort) types.CommonThinkingEffort {
	switch inputEffort {
	case types.EffortMedium, types.EffortHigh:
		return types.EffortHigh
	default:
		return types.EffortLow
	}
}
