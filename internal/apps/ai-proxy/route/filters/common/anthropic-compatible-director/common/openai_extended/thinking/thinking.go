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

// UnifiedThinking is the unified thinking configs.
// Distinguish:
// - not-set
// - enable
// - disable
type UnifiedThinking AnthropicThinking

func (t *UnifiedThinking) ToAnthropicThinking(maxTokens int) *AnthropicThinking {
	if t == nil || t.Thinking == nil {
		return nil
	}
	// check budget_tokens
	if t.Thinking.Type == "enabled" {
		suitableBudget := t.Thinking.BudgetTokens
		halfMaxTokens := maxTokens / 2
		if suitableBudget == 0 || suitableBudget >= maxTokens {
			suitableBudget = halfMaxTokens
		}
		if suitableBudget < 1024 {
			suitableBudget = 1024 // minimum budget, see: https://docs.anthropic.com/en/api/messages#body-thinking-budget-tokens
		}
		if suitableBudget >= maxTokens {
			panic("max_tokens should be greater than 1024 when you enable thinking")
		}
		t.Thinking.BudgetTokens = suitableBudget
	}
	return &AnthropicThinking{Thinking: t.Thinking}
}

func (t *UnifiedThinking) ToQwenThinking() *QwenThinking {
	if t == nil || t.Thinking == nil {
		return nil
	}
	enableThinking := t.Thinking.Type == "enabled"
	return &QwenThinking{EnableThinking: &enableThinking, ThinkingBudget: t.Thinking.BudgetTokens}
}

type (
	// Anthropic Thinking Style:
	//
	//	{
	//	  ...
	//	  "messages": ...,
	//	  "thinking": {
	//	    "type": "enabled",
	//	    "budget_tokens": 1000
	//	  }
	//	}
	AnthropicThinking struct {
		Thinking *AnthropicThinkingInternal `json:"thinking,omitempty"`
	}
	AnthropicThinkingInternal struct {
		Type         string `json:"type,omitempty"`          // enabled | disabled
		BudgetTokens int    `json:"budget_tokens,omitempty"` // should be small than max_tokens
	}
)

// Qwen Thinking Style:
//
//	{
//	  ...
//	  "messages": ...,
//	  "enable_thinking": true,
//	  "thinking_budget": 1000
//	}
type QwenThinking struct {
	EnableThinking *bool `json:"enable_thinking,omitempty"`
	ThinkingBudget int   `json:"thinking_budget,omitempty"`
}
