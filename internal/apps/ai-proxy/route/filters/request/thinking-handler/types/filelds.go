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

package types

// field constants
const (
	// for anthropic: https://docs.anthropic.com/en/api/messages#body-thinking
	// for bytedance: https://www.volcengine.com/docs/82379/1494384#:~:text=auto%EF%BC%9A-,%E8%87%AA%E5%8A%A8%E6%80%9D%E8%80%83%E6%A8%A1%E5%BC%8F,-%EF%BC%8C%E6%A8%A1%E5%9E%8B%E6%A0%B9%E6%8D%AE%E9%97%AE%E9%A2%98
	FieldThinking     = "thinking"
	FieldType         = "type"
	FieldBudgetTokens = "budget_tokens"

	// for qwen: https://help.aliyun.com/zh/model-studio/use-qwen-by-calling-api#:~:text=enable_thinking%22%3A%20xxx%7D%E3%80%82-,thinking_budget,-integer%20%EF%BC%88%E5%8F%AF%E9%80%89
	FieldEnableThinking = "enable_thinking"
	FieldThinkingBudget = "thinking_budget"

	// for openai responses: https://platform.openai.com/docs/api-reference/responses/create#responses_create-reasoning
	FieldReasoning = "reasoning"
	FieldEffort    = "effort"
	// for openai chat-completions: https://platform.openai.com/docs/api-reference/chat/create#chat_create-reasoning_effort
	FieldReasoningEffort = "reasoning_effort"

	// for vertex-ai: https://cloud.google.com/vertex-ai/generative-ai/docs/thinking
	FieldThinkingConfig = "thinking_config"
)
