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

// AnthropicResponse is non-stream response body.
// See: https://docs.anthropic.com/en/api/messages#response-id
type AnthropicResponse struct {
	ID      string                         `json:"id"`
	Type    string                         `json:"type"` // always "message"
	Role    string                         `json:"role"` // always "assistant"
	Content []AnthropicResponseContentPart `json:"content"`
	Usage   AnthropicResponseUsage         `json:"usage"`
}

type AnthropicResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
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
	for i, anthropicContentPart := range r.Content {
		anthropicContentPart := anthropicContentPart
		openaiMsg, err := ConvertAnthropicResponseMessageToOpenAIFormat(r.Role, anthropicContentPart)
		if err != nil {
			return openaiResp, fmt.Errorf("failed to convert anthropic response message to openai format: %v", err)
		}
		choices = append(choices, openai.ChatCompletionChoice{Message: openaiMsg})
		// Set finish reason for the last message
		if i == len(r.Content)-1 {
			choices[i].FinishReason = "stop"
		}
	}
	openaiResp.Choices = choices
	openaiResp.Usage = openai.Usage{
		PromptTokens:            r.Usage.InputTokens,
		CompletionTokens:        r.Usage.OutputTokens,
		TotalTokens:             r.Usage.InputTokens + r.Usage.OutputTokens,
		PromptTokensDetails:     nil,
		CompletionTokensDetails: nil,
	}
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

	Usage AnthropicResponseUsage
}

// ConvertStreamChunkDataToOpenAIChunk .
// See: https://docs.anthropic.com/en/docs/build-with-claude/streaming
func ConvertStreamChunkDataToOpenAIChunk(anthropicDataRaw []byte, inputMsgInfo AnthropicStreamMessageInfo) (*AnthropicStreamMessageInfo, *openai.ChatCompletionStreamResponse, error) {
	var rawObj map[string]any
	if err := json.Unmarshal(anthropicDataRaw, &rawObj); err != nil {
		return nil, nil, fmt.Errorf("failed to parse raw, err: %v", err)
	}

	switch rawObj["type"].(string) {
	case "message_start":
		// Example:
		// event: message_start
		// data: {"type": "message_start", "message": {"id": "msg_1nZdL29xx5MUA1yADyHTEsnR8uuvGzszyY", "type": "message", "role": "assistant", "content": [], "model": "claude-opus-4-20250514", "stop_reason": null, "stop_sequence": null, "usage": {"input_tokens": 25, "output_tokens": 1}}}

		// get message id and model
		var gotMsgInfo AnthropicStreamMessageInfo
		gotMsgInfo.ID = rawObj["message"].(map[string]any)["id"].(string)
		gotMsgInfo.Model = rawObj["message"].(map[string]any)["model"].(string)
		gotMsgInfo.Role = rawObj["message"].(map[string]any)["role"].(string)
		gotMsgInfo.Usage.InputTokens = int(rawObj["message"].(map[string]any)["usage"].(map[string]any)["input_tokens"].(float64))
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
	case "message_delta":
		// Example:
		// event: message_delta
		// data: {"type": "message_delta", "delta": {"stop_reason": "end_turn", "stop_sequence":null}, "usage": {"output_tokens": 15}}

		// Only send `finish_reason`; `usage` is handled in `message_stop`.

		outputTokens := int(rawObj["usage"].(map[string]any)["output_tokens"].(float64))
		inputMsgInfo.Usage.OutputTokens = outputTokens

		openaiChunk := openai.ChatCompletionStreamResponse{
			ID:      inputMsgInfo.ID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   inputMsgInfo.Model,
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index:        0,
					Delta:        openai.ChatCompletionStreamChoiceDelta{},
					FinishReason: "stop", // always stop
				},
			},
		}
		return &inputMsgInfo, &openaiChunk, nil
	case "message_stop":
		// Example:
		// event: message_stop
		// data: {"type": "message_stop"}

		openaiChunk := openai.ChatCompletionStreamResponse{
			ID:      inputMsgInfo.ID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   inputMsgInfo.Model,
			Choices: []openai.ChatCompletionStreamChoice{},
			Usage: &openai.Usage{
				PromptTokens:            inputMsgInfo.Usage.InputTokens,
				CompletionTokens:        inputMsgInfo.Usage.OutputTokens,
				TotalTokens:             inputMsgInfo.Usage.InputTokens + inputMsgInfo.Usage.OutputTokens,
				PromptTokensDetails:     nil,
				CompletionTokensDetails: nil,
			},
		}
		return nil, &openaiChunk, nil

	default:
		return nil, nil, nil
	}
}
