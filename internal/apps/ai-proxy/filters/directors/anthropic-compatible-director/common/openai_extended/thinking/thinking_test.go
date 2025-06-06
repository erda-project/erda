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

package thinking

import (
	"reflect"
	"testing"
)

func TestThinking_ToAnthropicThinking(t *testing.T) {
	type fields struct {
		UnifiedThinking *UnifiedThinking
	}
	tests := []struct {
		name   string
		fields fields
		want   *AnthropicThinking
	}{
		{
			name: "nil",
			fields: fields{
				UnifiedThinking: nil,
			},
			want: nil,
		},
		{
			name: "enabled",
			fields: fields{
				UnifiedThinking: &UnifiedThinking{
					Thinking: &AnthropicThinkingInternal{Type: "enabled", BudgetTokens: 100},
				},
			},
			want: &AnthropicThinking{
				Thinking: &AnthropicThinkingInternal{Type: "enabled", BudgetTokens: 100},
			},
		},
		{
			name: "disabled",
			fields: fields{
				UnifiedThinking: &UnifiedThinking{
					Thinking: &AnthropicThinkingInternal{Type: "disabled", BudgetTokens: 100},
				},
			},
			want: &AnthropicThinking{
				Thinking: &AnthropicThinkingInternal{Type: "disabled", BudgetTokens: 100},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.UnifiedThinking.ToAnthropicThinking(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Thinking.ToAnthropicThinking() = %v, want %v", got, tt.want)
			}
		})
	}
}
