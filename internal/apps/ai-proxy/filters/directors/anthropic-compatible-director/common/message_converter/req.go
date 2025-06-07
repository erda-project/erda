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
	"fmt"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/anthropic-compatible-director/common/openai_extended"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/anthropic-compatible-director/common/openai_extended/thinking"
)

const (
	defaultMaxTokens   = 10240
	defaultTemperature = 1.0
)

type BaseAnthropicRequest struct {
	Messages      []AnthropicMessage `json:"messages"`
	MaxTokens     int                `json:"max_tokens"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	System        string             `json:"system,omitempty"`
	Temperature   float32            `json:"temperature"`
	ToolChoice    any                `json:"tool_choice,omitempty"`
	Tools         any                `json:"tools,omitempty"`

	*thinking.AnthropicThinking
}

func ConvertOpenAIRequestToBaseAnthropicRequest(openaiReq openai_extended.OpenAIRequestExtended) BaseAnthropicRequest {
	// convert to: anthropic format request
	baseAnthropicReq := BaseAnthropicRequest{
		MaxTokens:     openaiReq.MaxTokens,
		Temperature:   openaiReq.Temperature,
		StopSequences: openaiReq.Stop,
	}
	if openaiReq.Temperature <= 0 {
		baseAnthropicReq.Temperature = defaultTemperature
	}
	if openaiReq.MaxTokens <= 1 {
		baseAnthropicReq.MaxTokens = defaultMaxTokens
	}
	// tool
	if len(openaiReq.Tools) > 0 {
		baseAnthropicReq.Tools = openaiReq.Tools
	}
	if openaiReq.ToolChoice != nil {
		baseAnthropicReq.ToolChoice = openaiReq.ToolChoice
	}
	// split system prompt out, keep user / assistant messages
	var systemPrompts []string
	for _, msg := range openaiReq.Messages {
		switch msg.Role {
		case openai.ChatMessageRoleSystem:
			systemPrompts = append(systemPrompts, msg.Content)
		default:
			bedrockMsg, err := ConvertOneOpenAIMessage(msg)
			if err != nil {
				panic(fmt.Errorf("failed to convert openai message to bedrock message: %v", err))
			}
			baseAnthropicReq.Messages = append(baseAnthropicReq.Messages, *bedrockMsg)
		}
	}
	if len(systemPrompts) > 0 {
		baseAnthropicReq.System = strings.Join(systemPrompts, "\n")
	}
	// thinking
	baseAnthropicReq.AnthropicThinking = thinking.UnifiedGetThinkingConfigs(openaiReq).ToAnthropicThinking()

	return baseAnthropicReq
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func ConvertOneOpenAIMessage(openaiMsg openai.ChatCompletionMessage) (*AnthropicMessage, error) {
	var anthropicMsg AnthropicMessage
	anthropicMsg.Role = openaiMsg.Role

	// only text content
	if openaiMsg.Content != "" {
		anthropicMsg.Content = openaiMsg.Content
		return &anthropicMsg, nil
	}

	var contentParts []map[string]any

	// multi content
	for _, part := range openaiMsg.MultiContent {
		switch part.Type {
		case openai.ChatMessagePartTypeText:
			bedrockPart := map[string]any{
				"type": "text",
				"text": part.Text,
			}
			contentParts = append(contentParts, bedrockPart)
		case openai.ChatMessagePartTypeImageURL:
			// get media_type and base64 content from url
			// url format: data:image/${media_type};base64,${base64_content}
			ss := strings.SplitN(part.ImageURL.URL, ";", 2)
			mediaType := strings.TrimPrefix(ss[0], "data:")
			base64Content := strings.TrimPrefix(ss[1], "base64,")
			bedrockPart := map[string]any{
				"type": "image",
				"source": map[string]any{
					"type":       "base64",
					"media_type": mediaType,
					"data":       base64Content,
				},
			}
			contentParts = append(contentParts, bedrockPart)
		}
	}

	anthropicMsg.Content = contentParts

	return &anthropicMsg, nil
}
