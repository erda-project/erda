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
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// QwenThinkingEncoder handles Qwen thinking encoding
type QwenThinkingEncoder struct{}

func (e *QwenThinkingEncoder) CanEncode(ctx context.Context) bool {
	model := ctxhelper.MustGetModel(ctx)
	return strings.EqualFold(model.Publisher, string(common_types.ModelPublisherQwen))
}

func (e *QwenThinkingEncoder) Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	if ct.MustGetMode() != types.ModeOn {
		return nil, nil
	}

	appendBodyMap := map[string]any{
		types.FieldEnableThinking: true,
	}

	// calculate suitable budget
	suitableBudget := 1024 // minimum budget for Anthropic
	// get from budget_tokens if provided
	if ct.BudgetTokens != nil && *ct.BudgetTokens > suitableBudget {
		suitableBudget = *ct.BudgetTokens
		goto RESULT
	}
	// get from effort if provided
	if ct.Effort != nil {
		switch *ct.Effort {
		case types.EffortMinimal:
			suitableBudget = 1024
		case types.EffortLow:
			suitableBudget = 2048
		case types.EffortMedium:
			suitableBudget = 4096
		case types.EffortHigh:
			suitableBudget = 8192
		}
		goto RESULT
	}

RESULT:
	appendBodyMap[types.FieldThinkingBudget] = suitableBudget
	return appendBodyMap, nil
}

func (e *QwenThinkingEncoder) GetPriority() int {
	return 2 // medium priority
}

func (e *QwenThinkingEncoder) GetName() string {
	return fmt.Sprintf("model_publisher: %s", common_types.ModelPublisherQwen)
}

var _ encoders.CommonThinkingEncoder = (*QwenThinkingEncoder)(nil)
