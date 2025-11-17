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

// OpenAIChatThinkingEncoder handles OpenAI Chat API thinking encoding
type OpenAIChatThinkingEncoder struct{}

func (e *OpenAIChatThinkingEncoder) CanEncode(ctx context.Context) bool {
	model := ctxhelper.MustGetModel(ctx)
	pathMatcher := ctxhelper.MustGetPathMatcher(ctx)

	return strings.EqualFold(model.Publisher, common_types.ModelPublisherOpenAI.String()) &&
		pathMatcher.Match(common.RequestPathPrefixV1ChatCompletions)
}

func (e *OpenAIChatThinkingEncoder) Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	if ct.MustGetMode() != types.ModeOn {
		return nil, nil
	}

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
	return map[string]any{types.FieldReasoningEffort: suitableEffort}, nil

}

func (e *OpenAIChatThinkingEncoder) GetPriority() int {
	return 3 // lower priority than single-provider encoders
}

func (e *OpenAIChatThinkingEncoder) GetName() string {
	return fmt.Sprintf("model_publisher: %s_chat", common_types.ModelPublisherOpenAI)
}

var _ encoders.CommonThinkingEncoder = (*OpenAIChatThinkingEncoder)(nil)
