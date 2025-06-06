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

package message_converter

import (
	"github.com/sashabaranov/go-openai"
)

// AnthropicResponseMessage is the message format of anthropic response.
// see: https://docs.anthropic.com/en/api/messages#response-content
type AnthropicResponseMessage struct {
	Role    string         `json:"role"`
	Content map[string]any `json:"content"`
}

func ConvertAnthropicResponseMessageToOpenAIFormat(anthropicMsg AnthropicResponseMessage) (openai.ChatCompletionMessage, error) {
	var openaiMsg openai.ChatCompletionMessage
	openaiMsg.Role = anthropicMsg.Role

	switch anthropicMsg.Content["type"].(string) {
	case "text":
		openaiMsg.Content = anthropicMsg.Content["text"].(string)
	case "thinking": // https://docs.anthropic.com/en/api/messages#thinking
		openaiMsg.ReasoningContent = anthropicMsg.Content["thinking"].(string)
	}

	return openaiMsg, nil
}
