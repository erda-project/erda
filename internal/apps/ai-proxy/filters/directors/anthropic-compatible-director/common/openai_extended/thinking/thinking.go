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

// Thinking is the unified thinking configs.
// Distinguish:
// - not-set
// - enable
// - disable
type Thinking AnthropicThinking

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
		Thinking *AnthropicThinkingInternal `json:"thinking"`
	}
	AnthropicThinkingInternal struct {
		Type         string `json:"type"`          // enabled | disabled
		BudgetTokens int    `json:"budget_tokens"` // should be small than max_tokens
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
	EnableThinking *bool `json:"enable_thinking"`
	ThinkingBudget int   `json:"thinking_budget"`
}
