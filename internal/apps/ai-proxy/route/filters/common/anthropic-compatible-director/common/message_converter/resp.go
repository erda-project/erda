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
	ID         string                         `json:"id"`
	Type       string                         `json:"type"` // always "message"
	Role       string                         `json:"role"` // always "assistant"
	Content    []AnthropicResponseContentPart `json:"content"`
	Usage      AnthropicResponseUsage         `json:"usage"`
	StopReason string                         `json:"stop_reason"`
}

type AnthropicResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (r AnthropicResponse) ConvertToOpenAIFormat(modelID string) (openai.ChatCompletionResponse, error) {
	openaiResp := openai.ChatCompletionResponse{
		ID:      r.ID,
		Object:  "chat.completion", // always
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
			choices[i].FinishReason = ConvertStopReason(r.StopReason)
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
	case "tool_use":
		// OpenAI Example:
		// "message": {
		//   "role": "assistant",
		//   "content": null,
		//   "tool_calls": [
		//     {
		//       "id": "call_kjMCowjrMJ55JVL6qLf3T5t5",
		//       "type": "function",
		//       "function": {
		//         "name": "get_weather",
		//         "arguments": "{\"location\":\"杭州, 中国\"}"
		//       }
		//     }
		//   ],
		//   "refusal": null,
		//   "annotations": []
		// },
		inputJSON, err := json.Marshal(anthropicContentPart["input"])
		if err != nil {
			return openaiMsg, fmt.Errorf("failed to marshal tool_use input JSON: %v", err)
		}
		openaiMsg.ToolCalls = []openai.ToolCall{
			{
				ID:   anthropicContentPart["id"].(string),
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      anthropicContentPart["name"].(string),
					Arguments: string(inputJSON),
				},
			},
		}
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

	case "content_block_start":
		// Example:
		// text:
		// event: content_block_start
		// data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}
		// tool:
		// event: content_block_start
		// data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_01T1x1fJ34qAmk2tNTrN7Up6","name":"get_weather","input":{}}}
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

		switch rawObj["content_block"].(map[string]any)["type"].(string) {
		case "text":
			// For text content block, we need to send a chunk with role but no content.
			// OpenAI Example:
			// data: {"id":"chatcmpl-Bx64y0qCiwYAxZTl6Zvap5s4M42GS","object":"chat.completion.chunk","created":1753424420,"model":"gpt-4.1-2025-04-14","system_fingerprint":"fp_07e970ab25","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}
			return nil, &openaiChunk, nil
		case "tool_use":
			// OpenAI Example:
			// data: {"id":"chatcmpl-Bx6B6JuBCCcfv6weh9OnZ7jjMgaou","object":"chat.completion.chunk","created":1753424800,"model":"gpt-4.1-2025-04-14","system_fingerprint":"fp_07e970ab25","choices":[{"index":0,"delta":{"role":"assistant","content":null,"tool_calls":[{"index":0,"id":"call_ChoPaqR61B15NXkGXXPFPQvZ","type":"function","function":{"name":"get_weather","arguments":""}}],"refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}
			openaiChunk.Choices[0].Delta.ToolCalls = []openai.ToolCall{
				{
					Index: &[]int{0}[0],
					ID:    rawObj["content_block"].(map[string]any)["id"].(string),
					Type:  "function", // always "function" for tool use
					Function: openai.FunctionCall{
						Name: rawObj["content_block"].(map[string]any)["name"].(string),
					},
				},
			}
			return nil, &openaiChunk, nil

		default:
			return nil, nil, nil
		}

	case "content_block_delta":
		// see: https://docs.anthropic.com/en/docs/build-with-claude/streaming#content-block-delta-types
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
		case "input_json_delta":
			// Anthropic Example:
			// event: content_block_delta
			// data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":""}}
			// OpenAI Example:
			// data: {"id":"chatcmpl-Bx6B6JuBCCcfv6weh9OnZ7jjMgaou","object":"chat.completion.chunk","created":1753424800,"model":"gpt-4.1-2025-04-14","system_fingerprint":"fp_07e970ab25","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"location"}}]},"logprobs":null,"finish_reason":null}],"usage":null}
			openaiChunk.Choices[0].Delta.ToolCalls = []openai.ToolCall{
				{
					Index: &[]int{0}[0],
					Function: openai.FunctionCall{
						Arguments: deltaObj["partial_json"].(string), // partial JSON string
					},
				},
			}
		default:
			return nil, nil, nil
		}
		return nil, &openaiChunk, nil

	case "content_block_stop":
		// Example:
		// event: content_block_stop
		// data: {"type":"content_block_stop","index":1}
		return nil, nil, nil

	case "message_delta":
		// Example:
		// event: message_delta
		// data: {"type": "message_delta", "delta": {"stop_reason": "end_turn", "stop_sequence":null}, "usage": {"output_tokens": 15}}

		// Only send `finish_reason`; `usage` is handled in `message_stop`.

		outputTokens := int(rawObj["usage"].(map[string]any)["output_tokens"].(float64))
		inputMsgInfo.Usage.OutputTokens = outputTokens

		stopReason := rawObj["delta"].(map[string]any)["stop_reason"].(string)

		openaiChunk := openai.ChatCompletionStreamResponse{
			ID:      inputMsgInfo.ID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   inputMsgInfo.Model,
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index:        0,
					Delta:        openai.ChatCompletionStreamChoiceDelta{},
					FinishReason: ConvertStopReason(stopReason),
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

func ConvertStopReason(stopReason string) openai.FinishReason {
	// anthropic: https://docs.anthropic.com/en/api/messages#response-stop-reason
	// openai: https://platform.openai.com/docs/api-reference/chat/object#chat/object-choices
	switch stopReason {
	case "tool_use":
		return openai.FinishReasonToolCalls
	case "refusal":
		return openai.FinishReasonContentFilter
	case "max_tokens":
		return openai.FinishReasonLength
	default:
		return openai.FinishReasonStop
	}
}
