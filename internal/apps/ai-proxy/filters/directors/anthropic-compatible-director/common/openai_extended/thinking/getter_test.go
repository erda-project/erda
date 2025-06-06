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

	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/anthropic-compatible-director/common/openai_extended"
)

func TestUnifiedGetThinkingConfigs(t *testing.T) {
	type args struct {
		req openai_extended.OpenAIRequestExtended
	}
	tests := []struct {
		name string
		args args
		want *Thinking
	}{
		{
			name: "anthropic style: enabled",
			args: args{
				req: openai_extended.OpenAIRequestExtended{
					ExtraFields: map[string]any{
						"thinking": map[string]any{
							"type":          "enabled",
							"budget_tokens": 100,
						},
					},
				},
			},
			want: &Thinking{
				Thinking: &AnthropicThinkingInternal{
					Type:         "enabled",
					BudgetTokens: 100,
				},
			},
		},
		{
			name: "anthropic style: disabled",
			args: args{
				req: openai_extended.OpenAIRequestExtended{
					ExtraFields: map[string]any{
						"thinking": map[string]any{
							"type":          "disabled",
							"budget_tokens": 100,
						},
					},
				},
			},
			want: &Thinking{
				Thinking: &AnthropicThinkingInternal{
					Type:         "disabled",
					BudgetTokens: 100,
				},
			},
		},
		{
			name: "qwen style: enabled",
			args: args{
				req: openai_extended.OpenAIRequestExtended{
					ExtraFields: map[string]any{
						"enable_thinking": true,
						"thinking_budget": 1000,
					},
				},
			},
			want: &Thinking{
				Thinking: &AnthropicThinkingInternal{
					Type:         "enabled",
					BudgetTokens: 1000,
				},
			},
		},
		{
			name: "qwen style: disabled",
			args: args{
				req: openai_extended.OpenAIRequestExtended{
					ExtraFields: map[string]any{
						"enable_thinking": false,
						"thinking_budget": 1000,
					},
				},
			},
			want: &Thinking{
				Thinking: &AnthropicThinkingInternal{
					Type:         "disabled",
					BudgetTokens: 1000,
				},
			},
		},
		{
			name: "not set: extra fields is empty",
			args: args{
				req: openai_extended.OpenAIRequestExtended{
					ExtraFields: map[string]any{},
				},
			},
			want: nil,
		},
		{
			name: "not set: extra fields is nil",
			args: args{
				req: openai_extended.OpenAIRequestExtended{
					ExtraFields: nil,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnifiedGetThinkingConfigs(tt.args.req); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnifiedGetThinkingConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
