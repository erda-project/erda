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
	"context"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// AnthropicThinkingEncoder handles Anthropic thinking encoding
type AnthropicThinkingEncoder struct{}

func (e *AnthropicThinkingEncoder) CanEncode(ctx context.Context) bool {
	model := ctxhelper.MustGetModel(ctx)
	return strings.EqualFold(model.Publisher, string(types.ModelPublisherAnthropic))
}

func (e *AnthropicThinkingEncoder) Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	thinkingObj := make(map[string]any)
	appendBodyMap := map[string]any{types.FieldThinking: thinkingObj}

	// treat auto as off
	if ct.MustGetMode() != types.ModeOn {
		thinkingObj[types.FieldType] = "disabled"
		return appendBodyMap, nil
	}

	thinkingObj[types.FieldType] = "enabled"

	// calculate suitable budget
	suitableBudget := 1024 // minimum budget for Anthropic
	// get from budget_tokens if provided
	if ct.BudgetTokens != nil && *ct.BudgetTokens > suitableBudget {
		suitableBudget = *ct.BudgetTokens
		goto RESULT
	}
	// get from effort if provided
	if ct.Effort != nil {
		suitableBudget = types.MapEffortToBudgetTokens(*ct.Effort)
		goto RESULT
	}
	// no suitable budget found, use default 1024
RESULT:
	thinkingObj[types.FieldBudgetTokens] = suitableBudget
	return appendBodyMap, nil
}

func (e *AnthropicThinkingEncoder) GetPriority() int {
	return 1 // high priority for anthropic
}

func (e *AnthropicThinkingEncoder) GetName() string {
	return "anthropic"
}
