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

package model_publisher

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// OpenAIResponsesThinkingEncoder handles OpenAI Responses API thinking encoding
type OpenAIResponsesThinkingEncoder struct{}

func (e *OpenAIResponsesThinkingEncoder) CanEncode(ctx context.Context) bool {
	model := ctxhelper.MustGetModel(ctx)
	pathMatcher := ctxhelper.MustGetPathMatcher(ctx)

	return strings.EqualFold(model.Publisher, string(common_types.ModelPublisherOpenAI)) &&
		pathMatcher.Match(common.RequestPathPrefixV1Responses)
}

func (e *OpenAIResponsesThinkingEncoder) Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	if ct.MustGetMode() != types.ModeOn {
		return nil, nil
	}

	reasoningObj := map[string]any{"summary": "auto"}
	appendBodyMap := map[string]any{types.FieldReasoning: reasoningObj}

	suitableEffort := types.EffortMedium
	// get from effort if provided
	if ct.Effort != nil {
		suitableEffort = *ct.Effort
		goto RESULT
	}
	// get from budget_tokens if provided
	if ct.BudgetTokens != nil && *ct.BudgetTokens > 0 {
		suitableEffort = types.MapBudgetTokensToEffort(*ct.BudgetTokens)
		goto RESULT
	}

RESULT:
	reasoningObj[types.FieldEffort] = suitableEffort
	return appendBodyMap, nil
}

func (e *OpenAIResponsesThinkingEncoder) GetPriority() int {
	return 4 // lower priority than chat encoder
}

func (e *OpenAIResponsesThinkingEncoder) GetName() string {
	return fmt.Sprintf("model_publisher: %s_responses", common_types.ModelPublisherOpenAI)
}

var _ encoders.CommonThinkingEncoder = (*OpenAIResponsesThinkingEncoder)(nil)
