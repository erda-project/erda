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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

func TestOpenAIChatThinkingExtractor_ExtractMode(t *testing.T) {
	extractor := &OpenAIChatThinkingExtractor{}
	tests := []struct {
		name     string
		body     map[string]any
		expected *types.CommonThinkingMode
	}{
		{
			name: "nil body",
			body: nil,
		},
		{
			name: "missing reasoning effort",
			body: map[string]any{
				"foo": "bar",
			},
			expected: nil,
		},
		{
			name: "invalid reasoning effort type",
			body: map[string]any{
				types.FieldReasoningEffort: 123,
			},
			expected: nil,
		},
		{
			name: "invalid reasoning effort value",
			body: map[string]any{
				types.FieldReasoningEffort: "invalid",
			},
			expected: nil,
		},
		{
			name: "none effort maps to mode off",
			body: map[string]any{
				types.FieldReasoningEffort: types.EffortNone.String(),
			},
			expected: types.ModePtr(types.ModeOff),
		},
		{
			name: "valid effort turns mode on",
			body: map[string]any{
				types.FieldReasoningEffort: types.EffortHigh.String(),
			},
			expected: types.ModePtr(types.ModeOn),
		},
		{
			name: "minimal effort also turns mode on",
			body: map[string]any{
				types.FieldReasoningEffort: types.EffortMinimal.String(),
			},
			expected: types.ModePtr(types.ModeOn),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.ExtractMode(tt.body)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenAIChatThinkingExtractor_ExtractEffort(t *testing.T) {
	extractor := &OpenAIChatThinkingExtractor{}
	tests := []struct {
		name     string
		body     map[string]any
		expected *types.CommonThinkingEffort
	}{
		{
			name: "missing reasoning effort",
			body: map[string]any{},
		},
		{
			name: "invalid reasoning effort type",
			body: map[string]any{
				types.FieldReasoningEffort: true,
			},
		},
		{
			name: "invalid reasoning effort value",
			body: map[string]any{
				types.FieldReasoningEffort: "invalid",
			},
		},
		{
			name: "valid reasoning effort returns pointer",
			body: map[string]any{
				types.FieldReasoningEffort: types.EffortLow.String(),
			},
			expected: types.EffortPtr(types.EffortLow),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.ExtractEffort(tt.body)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenAIChatThinkingExtractor_Metadata(t *testing.T) {
	extractor := &OpenAIChatThinkingExtractor{}
	assert.Equal(t, 4, extractor.GetPriority())
	assert.Equal(t, "openai_chat", extractor.GetName())
}

func TestOpenAIChatThinkingExtractor_RelatedBehavior(t *testing.T) {
	extractor := &OpenAIChatThinkingExtractor{}

	body := map[string]any{
		types.FieldReasoningEffort: types.EffortMedium.String(),
		"other":                    1,
	}

	assert.True(t, extractor.CanExtract(body))
	assert.False(t, extractor.CanExtract(map[string]any{}))

	related := extractor.RelatedFields(body)
	require.Len(t, related, 1)
	assert.Equal(t, types.EffortMedium.String(), related[types.FieldReasoningEffort])

	budget, err := extractor.ExtractBudgetTokens(body)
	require.NoError(t, err)
	assert.Nil(t, budget)
}
