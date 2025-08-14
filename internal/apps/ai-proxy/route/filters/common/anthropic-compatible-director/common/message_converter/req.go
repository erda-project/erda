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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/common/openai_extended"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/common/openai_extended/thinking"
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
		var aTools []AnthropicTool
		for _, oTool := range openaiReq.Tools {
			aTool, err := ConvertOneOpenAITool(oTool)
			if err != nil {
				panic(fmt.Errorf("failed to convert openai tool to anthropic tool: %v, oTool: %#v", err, oTool))
			}
			aTools = append(aTools, *aTool)
		}
		baseAnthropicReq.Tools = aTools
	}
	if openaiReq.ToolChoice != nil {
		baseAnthropicReq.ToolChoice = openaiReq.ToolChoice
	}
	// split system prompt out, keep user / assistant messages
	var systemPrompts []string
	for i, msg := range openaiReq.Messages {
		switch msg.Role {
		case openai.ChatMessageRoleSystem:
			systemPrompts = append(systemPrompts, msg.Content)
		case openai.ChatMessageRoleUser:
			bedrockMsg, err := ConvertOneOpenAIUserMessage(msg)
			if err != nil || bedrockMsg.Content == nil {
				panic(fmt.Errorf("failed to convert openai user message to bedrock message: %v, index: %d, msg: %#v", err, i, msg))
			}
			baseAnthropicReq.Messages = append(baseAnthropicReq.Messages, *bedrockMsg)
		case openai.ChatMessageRoleTool:
			bedrockMsg, err := ConvertOneOpenAIToolMessage(msg)
			if err != nil {
				panic(fmt.Errorf("failed to convert openai tool message to bedrock message: %v, index: %d, msg: %#v", err, i, msg))
			}
			baseAnthropicReq.Messages = append(baseAnthropicReq.Messages, *bedrockMsg)
		case openai.ChatMessageRoleAssistant:
			bedrockMsg, err := ConvertOneOpenAIAssistantMessage(msg)
			if err != nil {
				panic(fmt.Errorf("failed to convert openai assistant message to bedrock message: %v, index: %d, msg: %#v", err, i, msg))
			}
			baseAnthropicReq.Messages = append(baseAnthropicReq.Messages, *bedrockMsg)
		default:
			panic(fmt.Errorf("unsupported msg role: %v, message index: %d", msg.Role, i))
		}
	}
	if len(systemPrompts) > 0 {
		baseAnthropicReq.System = strings.Join(systemPrompts, "\n")
	}
	// thinking
	baseAnthropicReq.AnthropicThinking = thinking.UnifiedGetThinkingConfigs(openaiReq).ToAnthropicThinking()

	// split anthropic messages
	splitAnthropicMessages(&baseAnthropicReq)

	return baseAnthropicReq
}

type AnthropicMessage struct {
	Role    string           `json:"role"`
	Content []map[string]any `json:"content"`
}

func ConvertOneOpenAIUserMessage(openaiMsg openai.ChatCompletionMessage) (*AnthropicMessage, error) {
	var anthropicMsg AnthropicMessage
	anthropicMsg.Role = openai.ChatMessageRoleUser

	// only text content
	if openaiMsg.Content != "" {
		anthropicMsg.Content = []map[string]any{
			{
				"type": "text",
				"text": openaiMsg.Content,
			},
		}
	}

	// multi content
	anthropicMsg.Content = append(anthropicMsg.Content, convertOpenAIMultiContent(openaiMsg.MultiContent)...)

	return &anthropicMsg, nil
}

func ConvertOneOpenAIToolMessage(openaiMsg openai.ChatCompletionMessage) (*AnthropicMessage, error) {
	var anthropicMsg AnthropicMessage
	anthropicMsg.Role = openai.ChatMessageRoleUser

	anthropicMsg.Content = []map[string]any{
		{
			"type":        "tool_result",
			"tool_use_id": openaiMsg.ToolCallID,
			"content":     openaiMsg.Content,
		},
	}

	return &anthropicMsg, nil
}

func ConvertOneOpenAIAssistantMessage(openaiMsg openai.ChatCompletionMessage) (*AnthropicMessage, error) {
	var anthropicMsg AnthropicMessage
	anthropicMsg.Role = openai.ChatMessageRoleAssistant

	// only text content
	if openaiMsg.Content != "" {
		anthropicMsg.Content = []map[string]any{
			{
				"type": "text",
				"text": openaiMsg.Content,
			},
		}
	}

	// multi content
	anthropicMsg.Content = append(anthropicMsg.Content, convertOpenAIMultiContent(openaiMsg.MultiContent)...)

	// tool_use
	for _, toolCall := range openaiMsg.ToolCalls {
		anthropicMsg.Content = append(anthropicMsg.Content, map[string]any{
			"type": "tool_use",
			"id":   toolCall.ID,
			"name": toolCall.Function.Name,
			"input": func() map[string]any {
				var input map[string]any
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					panic(fmt.Errorf("failed to unmarshal tool call arguments: %v, arguments: %s", err, toolCall.Function.Arguments))
				}
				return input
			}(),
		})
	}

	return &anthropicMsg, nil
}

func convertOpenAIMultiContent(openaiMultiContent []openai.ChatMessagePart) []map[string]any {
	var anthropicContent []map[string]any
	for _, part := range openaiMultiContent {
		switch part.Type {
		case openai.ChatMessagePartTypeText:
			bedrockPart := map[string]any{
				"type": "text",
				"text": part.Text,
			}
			anthropicContent = append(anthropicContent, bedrockPart)
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
			anthropicContent = append(anthropicContent, bedrockPart)
		}
	}
	return anthropicContent
}

type AnthropicTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

func ConvertOneOpenAITool(openaiTool openai.Tool) (*AnthropicTool, error) {
	if openaiTool.Type != openai.ToolTypeFunction {
		return nil, fmt.Errorf("unsupported tool type: %s", openaiTool.Type)
	}
	if openaiTool.Function == nil {
		return nil, fmt.Errorf("tool function is nil")
	}
	aTool := AnthropicTool{
		Name:        openaiTool.Function.Name,
		Description: openaiTool.Function.Description,
		InputSchema: openaiTool.Function.Parameters,
	}
	return &aTool, nil
}

// splitAnthropicMessages split some message into multiples.
// For Example:
// -> tips: 'image' blocks are not permitted within assistant turns.
// so we have to change this assistant image content part into a user message.
func splitAnthropicMessages(req *BaseAnthropicRequest) {
	var newMessages []AnthropicMessage

	for _, oldMessage := range req.Messages {
		if oldMessage.Role != openai.ChatMessageRoleAssistant {
			newMessages = append(newMessages, oldMessage)
			continue
		}

		var textBuf []map[string]any  // accumulate non‑image parts
		var imageBuf []map[string]any // accumulate consecutive image parts

		flushText := func() {
			if len(textBuf) == 0 {
				return
			}
			newMessages = append(newMessages, AnthropicMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: textBuf,
			})
			textBuf = nil
		}
		flushImage := func() {
			if len(imageBuf) == 0 {
				return
			}
			newMessages = append(newMessages, AnthropicMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: imageBuf,
			})
			imageBuf = nil
		}

		for _, part := range oldMessage.Content {
			if part["type"] == "image" {
				// Image block: flush previous text first, then accumulate image
				flushText()
				imageBuf = append(imageBuf, part)
				continue
			}
			// Non-image block: flush previous images first, then accumulate text
			flushImage()
			textBuf = append(textBuf, part)
		}
		// flush any remaining buffers
		flushText()
		flushImage()
	}
	// ------------------------------------------------------------------
	// Ensure the final role alternates correctly:
	// If the original last message was assistant *but* splitting makes
	// the new last message user (image), append a dummy assistant reply.
	// ------------------------------------------------------------------
	if len(req.Messages) > 0 && len(newMessages) > 0 {
		origLastRole := req.Messages[len(req.Messages)-1].Role
		newLastRole := newMessages[len(newMessages)-1].Role

		if origLastRole == openai.ChatMessageRoleAssistant &&
			newLastRole == openai.ChatMessageRoleUser {
			newMessages = append(newMessages, AnthropicMessage{
				Role: openai.ChatMessageRoleAssistant,
				Content: []map[string]any{
					{
						"type": "text",
						"text": "copy that",
					},
				},
			})
		}
	}

	req.Messages = newMessages
}
