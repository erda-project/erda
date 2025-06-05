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

package anthropic_official

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	APIVendor api_style.APIVendor = "anthropic"

	defaultMaxTokens      = 1024
	defaultTemperature    = 1.0
	defaultBedrockVersion = "bedrock-2023-05-31"
)

type AnthropicRequest struct {
	Model         string             `json:"model"`
	Messages      []AnthropicMessage `json:"messages"`
	MaxTokens     int                `json:"max_tokens"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	Stream        bool               `json:"stream"`
	System        string             `json:"system,omitempty"`
	Temperature   float32            `json:"temperature"`
	ToolChoice    any                `json:"tool_choice,omitempty"`
	Tools         any                `json:"tools,omitempty"`
	TopP          float32            `json:"top_p,omitempty"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"` // always "message"
	Role         string           `json:"role"` // always "assistant"
	Model        string           `json:"model"`
	Content      []map[string]any `json:"content"`
	StopReason   string           `json:"stop_reason"`
	StopSequence string           `json:"stop_sequence,omitempty"`
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
	for _, contentPart := range r.Content {
		switch contentPart["type"].(string) {
		case "text":
			choice := openai.ChatCompletionChoice{
				Message: openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: contentPart["text"].(string),
				}}
			choices = append(choices, choice)
		default:
			return openaiResp, fmt.Errorf("unsupported content type: %s", contentPart["type"].(string))
		}
	}
	openaiResp.Choices = choices
	return openaiResp, nil
}

type AnthropicDirector struct {
	*reverseproxy.DefaultResponseFilter

	StreamMessageID    string
	StreamMessageModel string
	StreamMessageRole  string
}

func NewDirector() *AnthropicDirector {
	return &AnthropicDirector{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}
}

func (f *AnthropicDirector) OfficialDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	model := ctxhelper.MustGetModel(ctx)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		// openai format request
		var openaiReq openai.ChatCompletionRequest
		if err := json.NewDecoder(infor.Body()).Decode(&openaiReq); err != nil {
			panic(fmt.Errorf("failed to decode request body as openai format, err: %v", err))
		}
		// convert to: anthropic format request
		anthropicReq := AnthropicRequest{
			Model:         model.Metadata.Public["model_name"].GetStringValue(),
			MaxTokens:     openaiReq.MaxTokens,
			StopSequences: openaiReq.Stop,
			Stream:        openaiReq.Stream,
			Temperature:   openaiReq.Temperature,
			TopP:          openaiReq.TopP,
		}
		if openaiReq.Temperature <= 0 {
			anthropicReq.Temperature = defaultTemperature
		}
		if openaiReq.MaxTokens <= 1 {
			anthropicReq.MaxTokens = defaultMaxTokens
		}
		// tool
		if len(openaiReq.Tools) > 0 {
			anthropicReq.Tools = openaiReq.Tools
		}
		if openaiReq.ToolChoice != nil {
			anthropicReq.ToolChoice = openaiReq.ToolChoice
		}
		// split system prompt out, keep user / assistant messages
		var systemPrompts []string
		for _, msg := range openaiReq.Messages {
			switch msg.Role {
			case openai.ChatMessageRoleSystem:
				systemPrompts = append(systemPrompts, msg.Content)
			default:
				anthropicReq.Messages = append(anthropicReq.Messages, AnthropicMessage{Role: msg.Role, Content: msg.Content})
			}
		}
		if len(systemPrompts) > 0 {
			anthropicReq.System = strings.Join(systemPrompts, "\n")
		}
		anthropicReqBytes, err := json.Marshal(&anthropicReq)
		if err != nil {
			panic(fmt.Errorf("failed to marshal anthropic request: %v", err))
		}
		infor.SetBody2(anthropicReqBytes)
	})

	return nil
}

func (f *AnthropicDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	// non-stream
	if !ctxhelper.GetIsStream(ctx) {
		f.DefaultResponseFilter.Buffer.Write(chunk) // write to buffer, so we can get allChunks later
		return reverseproxy.Continue, nil
	}
	// stream
	var chunkWriter bytes.Buffer
	openaiChunks, err := f.pipeAnthropicStream(ctx, io.NopCloser(bytes.NewBuffer(chunk)), &chunkWriter)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to parse anthropic eventstream, err: %v", err)
	}
	for _, openaiChunk := range openaiChunks {
		b, err := json.Marshal(openaiChunk)
		if err != nil {
			return reverseproxy.Intercept, fmt.Errorf("failed to marshal openai chunk, err: %v", err)
		}
		chunkData := vars.ConcatChunkDataPrefix(b)
		if _, err := w.Write(chunkData); err != nil {
			return reverseproxy.Intercept, fmt.Errorf("failed to write openai chunk, err: %v", err)
		}
	}
	return reverseproxy.Continue, nil
}

func (f *AnthropicDirector) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (err error) {
	// only stream style need append [DONE] chunk
	if !ctxhelper.GetIsStream(ctx) {
		// convert all at once
		var anthropicResp AnthropicResponse
		if err := json.Unmarshal(f.DefaultResponseFilter.Buffer.Bytes(), &anthropicResp); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %s, err: %v", string(chunk), err)
		}
		openaiResp, err := anthropicResp.ConvertToOpenAIFormat(ctxhelper.MustGetModel(ctx).Metadata.Public["model_id"].GetStringValue())
		if err != nil {
			return fmt.Errorf("failed to convert anthropic response body to openai format, err: %v", err)
		}
		b, err := json.Marshal(openaiResp)
		if err != nil {
			return fmt.Errorf("failed to marshal openai resp, err: %v", err)
		}
		infor.Header().Del("Content-Length")
		return f.DefaultResponseFilter.OnResponseEOF(ctx, nil, w, b)
	}
	// append [DONE] chunk
	doneChunk := vars.ConcatChunkDataPrefix([]byte("[DONE]"))
	if _, err := w.Write(doneChunk); err != nil {
		return fmt.Errorf("failed to write openai chunk, err: %v", err)
	}
	return nil
}
