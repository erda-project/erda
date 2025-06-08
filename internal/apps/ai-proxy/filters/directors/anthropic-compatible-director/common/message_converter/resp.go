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
	"time"

	"github.com/sashabaranov/go-openai"
)

type AnthropicResponse struct {
	ID      string                         `json:"id"`
	Type    string                         `json:"type"` // always "message"
	Role    string                         `json:"role"` // always "assistant"
	Content []AnthropicResponseContentPart `json:"content"`
}

func (r AnthropicResponse) ConvertToOpenAIFormat(modelID string) (openai.ChatCompletionResponse, error) {
	openaiResp := openai.ChatCompletionResponse{
		ID:      r.ID,
		Object:  "chat.completions", // always
		Created: time.Now().Unix(),
		Model:   modelID,
	}
	// choices
	var choices []openai.ChatCompletionChoice
	for _, anthropicContentPart := range r.Content {
		anthropicContentPart := anthropicContentPart
		openaiMsg, err := ConvertAnthropicResponseMessageToOpenAIFormat(r.Role, anthropicContentPart)
		if err != nil {
			return openaiResp, fmt.Errorf("failed to convert anthropic response message to openai format: %v", err)
		}
		choices = append(choices, openai.ChatCompletionChoice{Message: openaiMsg})
	}
	openaiResp.Choices = choices
	return openaiResp, nil
}

// AnthropicResponseContentPart is the message format of anthropic response.
// see: https://docs.anthropic.com/en/api/messages#response-content
type AnthropicResponseContentPart map[string]any

func ConvertAnthropicResponseMessageToOpenAIFormat(role string, anthropicContentPart AnthropicResponseContentPart) (openai.ChatCompletionMessage, error) {
	var openaiMsg openai.ChatCompletionMessage
	openaiMsg.Role = role

	switch anthropicContentPart["type"].(string) {
	case "text":
		openaiMsg.Content = anthropicContentPart["text"].(string)
	case "thinking": // https://docs.anthropic.com/en/api/messages#thinking
		openaiMsg.ReasoningContent = anthropicContentPart["thinking"].(string)
	}

	return openaiMsg, nil
}

type AnthropicStreamMessageInfo struct {
	ID    string
	Model string
	Role  string
}

func ConvertStreamChunkDataToOpenAIChunk(anthropicDataRaw []byte, inputMsgInfo AnthropicStreamMessageInfo) (*AnthropicStreamMessageInfo, *openai.ChatCompletionStreamResponse, error) {
	var rawObj map[string]any
	if err := json.Unmarshal(anthropicDataRaw, &rawObj); err != nil {
		return nil, nil, fmt.Errorf("failed to parse raw, err: %v", err)
	}

	switch rawObj["type"].(string) {
	case "message_start":
		// get message id and model
		var gotMsgInfo AnthropicStreamMessageInfo
		gotMsgInfo.ID = rawObj["message"].(map[string]any)["id"].(string)
		gotMsgInfo.Model = rawObj["message"].(map[string]any)["model"].(string)
		gotMsgInfo.Role = rawObj["message"].(map[string]any)["role"].(string)
		return &gotMsgInfo, nil, nil
	case "content_block_delta":
		openaiChunk := openai.ChatCompletionStreamResponse{
			ID:      inputMsgInfo.ID,
			Object:  "chat.completion.chunk", // fixed
			Created: time.Now().Unix(),
			Model:   inputMsgInfo.Model,
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Role: inputMsgInfo.Role,
					},
				},
			},
		}
		// get delta
		deltaObj := rawObj["delta"].(map[string]any)
		switch deltaObj["type"].(string) {
		case "text_delta":
			openaiChunk.Choices[0].Delta.Content = deltaObj["text"].(string)
		case "thinking_delta":
			openaiChunk.Choices[0].Delta.ReasoningContent = deltaObj["thinking"].(string)
		default:
			return nil, nil, nil
		}
		return nil, &openaiChunk, nil

	default:
		return nil, nil, nil
	}
}
