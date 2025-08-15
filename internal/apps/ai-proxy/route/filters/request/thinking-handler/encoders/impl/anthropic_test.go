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

func TestAnthropicThinkingEncoder_Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    types.CommonThinking
		expected map[string]any
	}{
		{
			name: "mode off - should return disabled",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeOff),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "disabled",
				},
			},
		},
		{
			name: "mode off with effort and budget - should ignore them",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOff),
				Effort:       types.EffortPtr(types.EffortHigh),
				BudgetTokens: intPtr(4096),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "disabled",
				},
			},
		},
		{
			name: "mode on with budget tokens > 1024",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(4096),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType:         "enabled",
					types.FieldBudgetTokens: 4096,
				},
			},
		},
		{
			name: "mode auto - should return disabled",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeAuto),
				BudgetTokens: intPtr(8192),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "disabled",
				},
			},
		},
		{
			name: "mode on with budget tokens <= 1024 - use 1024",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				BudgetTokens: intPtr(512),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType:         "enabled",
					types.FieldBudgetTokens: 1024,
				},
			},
		},
		{
			name: "mode on with effort minimal",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortMinimal),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType:         "enabled",
					types.FieldBudgetTokens: 1024,
				},
			},
		},
		{
			name: "mode on with effort low",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortLow),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType:         "enabled",
					types.FieldBudgetTokens: 2048,
				},
			},
		},
		{
			name: "mode on with effort medium",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortMedium),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType:         "enabled",
					types.FieldBudgetTokens: 4096,
				},
			},
		},
		{
			name: "mode on with effort high",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeOn),
				Effort: types.EffortPtr(types.EffortHigh),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType:         "enabled",
					types.FieldBudgetTokens: 8192,
				},
			},
		},
		{
			name: "mode auto with effort - should return disabled",
			input: types.CommonThinking{
				Mode:   types.ModePtr(types.ModeAuto),
				Effort: types.EffortPtr(types.EffortMedium),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "disabled",
				},
			},
		},
		{
			name: "mode on without effort or budget - use default 1024",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeOn),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType:         "enabled",
					types.FieldBudgetTokens: 1024,
				},
			},
		},
		{
			name: "mode auto without effort or budget - should return disabled",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeAuto),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "disabled",
				},
			},
		},
	}

	encoder := &AnthropicThinkingEncoder{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encoder.Encode(ctx, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// helper function for tests
func intPtr(i int) *int {
	return &i
}
