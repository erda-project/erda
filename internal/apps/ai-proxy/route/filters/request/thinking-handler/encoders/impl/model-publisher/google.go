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

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// GoogleThinkingEncoder handles Google thinking encoding
type GoogleThinkingEncoder struct{}

func (e *GoogleThinkingEncoder) CanEncode(ctx context.Context) bool {
	model := ctxhelper.MustGetModel(ctx)
	return strings.EqualFold(model.Publisher, string(common_types.ModelPublisherGoogle))
}

func (e *GoogleThinkingEncoder) Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	if ct.MustGetMode() == types.ModeOff {
		return nil, nil
	}
	thinkingObj := map[string]any{
		"include_thoughts": true,
	}
	appendBodyMap := map[string]any{types.FieldThinkingConfig: thinkingObj}

	// calculate suitable budget
	suitableBudget := 1024
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
	thinkingObj[types.FieldThinkingBudget] = suitableBudget
	return appendBodyMap, nil
}

func (e *GoogleThinkingEncoder) GetPriority() int {
	return 5
}

func (e *GoogleThinkingEncoder) GetName() string {
	return fmt.Sprintf("model_publisher: %s", common_types.ModelPublisherGoogle)
}
