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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

func TestQwenThinkingEncoder_Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    types.CommonThinking
		expected map[string]any
	}{
		{
			name: "mode off - should return nil",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeOff),
			},
			expected: nil,
		},
		{
			name: "mode auto - should return nil",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeAuto),
			},
			expected: nil,
		},
		{
			name: "mode on with budget tokens > 1024",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(4096),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 4096,
			},
		},
		{
			name: "mode on with budget tokens <= 1024 - use provided value",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(512),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 1024, // uses default minimum
			},
		},
		{
			name: "mode on with effort minimal",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortMinimal),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 1024,
			},
		},
		{
			name: "mode on with effort low",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortLow),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 2048,
			},
		},
		{
			name: "mode on with effort medium",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortMedium),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 4096,
			},
		},
		{
			name: "mode on with effort high",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortHigh),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 8192,
			},
		},
		{
			name: "mode on without effort or budget - use default 1024",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeOn),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 1024,
			},
		},
		{
			name: "mode on with both effort and budget - budget takes priority",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				Effort:       types.EffortPtr(types.EffortHigh), // would map to 8192
				BudgetTokens: intPtr(2048),                      // budget should win
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 2048,
			},
		},
		{
			name: "no mode with effort - should default to on and use effort",
			input: types.CommonThinking{
				Effort: types.EffortPtr(types.EffortMedium),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 4096,
			},
		},
		{
			name: "no mode with budget - should default to on and use budget",
			input: types.CommonThinking{
				BudgetTokens: intPtr(3000),
			},
			expected: map[string]any{
				types.FieldEnableThinking: true,
				types.FieldThinkingBudget: 3000,
			},
		},
	}

	encoder := &QwenThinkingEncoder{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encoder.Encode(ctx, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
