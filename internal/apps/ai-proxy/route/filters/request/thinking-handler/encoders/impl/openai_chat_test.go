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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

func TestOpenAIChatThinkingEncoder_Encode(t *testing.T) {
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
			name: "mode on with explicit effort minimal",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortMinimal),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortMinimal,
			},
		},
		{
			name: "mode on with explicit effort low",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortLow),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortLow,
			},
		},
		{
			name: "mode on with explicit effort medium",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortMedium),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortMedium,
			},
		},
		{
			name: "mode on with explicit effort high",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortHigh),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortHigh,
			},
		},
		{
			name: "mode on with budget < 1024 - maps to minimal",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(512),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortMinimal,
			},
		},
		{
			name: "mode on with budget 1024-2047 - maps to low",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(1500),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortLow,
			},
		},
		{
			name: "mode on with budget 2048-4095 - maps to medium",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(3000),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortMedium,
			},
		},
		{
			name: "mode on with budget >= 4096 - maps to high",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(8192),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortHigh,
			},
		},
		{
			name: "mode on with budget 0 - should use default medium",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(0),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortMedium,
			},
		},
		{
			name: "mode on without effort or budget - use default medium",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeOn),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortMedium,
			},
		},
		{
			name: "mode on with both effort and budget - effort takes priority",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				Effort:       types.EffortPtr(types.EffortLow),
				BudgetTokens: intPtr(8192), // would map to high, but effort should win
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortLow,
			},
		},
		{
			name: "no mode with effort - should default to on and use effort",
			input: types.CommonThinking{
				Effort: types.EffortPtr(types.EffortHigh),
			},
			expected: map[string]any{
				types.FieldReasoningEffort: types.EffortHigh,
			},
		},
	}

	encoder := &OpenAIChatThinkingEncoder{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encoder.Encode(ctx, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

