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

func TestBytedanceThinkingEncoder_Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    types.CommonThinking
		expected map[string]any
	}{
		{
			name: "mode on",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeOn),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "enabled",
				},
			},
		},
		{
			name: "mode off",
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
			name: "mode auto",
			input: types.CommonThinking{
				Mode: types.ModePtr(types.ModeAuto),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "auto",
				},
			},
		},
		{
			name: "mode on with effort and budget - should ignore them",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeOn),
				Effort:       types.EffortPtr(types.EffortHigh),
				BudgetTokens: intPtr(4096),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "enabled",
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
			name: "mode auto with effort and budget - should ignore them",
			input: types.CommonThinking{
				Mode:         types.ModePtr(types.ModeAuto),
				Effort:       types.EffortPtr(types.EffortMedium),
				BudgetTokens: intPtr(2048),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "auto",
				},
			},
		},
		{
			name: "no mode - should default based on MustGetMode logic",
			input: types.CommonThinking{
				Effort: types.EffortPtr(types.EffortHigh),
			},
			expected: map[string]any{
				types.FieldThinking: map[string]any{
					types.FieldType: "enabled", // MustGetMode should return ModeOn when effort is present
				},
			},
		},
	}

	encoder := &BytedanceThinkingEncoder{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encoder.Encode(ctx, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

