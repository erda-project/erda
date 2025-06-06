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

package aws_bedrock

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/smithy-go/logging"
	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/anthropic-compatible-director/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	APIVendor api_style.APIVendor = "aws-bedrock"

	defaultMaxTokens      = 1024
	defaultTemperature    = 1.0
	defaultBedrockVersion = "bedrock-2023-05-31"
)

type BedrockRequest struct {
	System           string                    `json:"system,omitempty"`
	Messages         []common.AnthropicMessage `json:"messages"`
	MaxTokens        int                       `json:"max_tokens"`
	Temperature      float32                   `json:"temperature"`
	TopP             float32                   `json:"top_p,omitempty"`
	Tools            any                       `json:"tools,omitempty"`
	ToolChoice       any                       `json:"tool_choice,omitempty"`
	StopSequences    []string                  `json:"stop_sequences,omitempty"`
	AnthropicVersion string                    `json:"anthropic_version"`
}

type BedrockResponse struct {
	ID      string           `json:"id"`
	Type    string           `json:"type"` // always "message"
	Role    string           `json:"role"` // always "assistant"
	Content []map[string]any `json:"content"`
}

func (r BedrockResponse) ConvertToOpenAIFormat(modelID string) (openai.ChatCompletionResponse, error) {
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

type BedrockDirector struct {
	*reverseproxy.DefaultResponseFilter

	StreamMessageID    string
	StreamMessageModel string
	StreamMessageRole  string
}

func NewDirector() *BedrockDirector {
	return &BedrockDirector{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}
}

func (f *BedrockDirector) AwsBedrockDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		// handle path for stream
		if ctxhelper.GetIsStream(ctx) {
			req.URL.Path = strings.ReplaceAll(req.URL.Path, "/invoke", "/invoke-with-response-stream")
		}
		// openai format request
		var openaiReq openai.ChatCompletionRequest
		if err := json.NewDecoder(infor.Body()).Decode(&openaiReq); err != nil {
			panic(fmt.Errorf("failed to decode request body as openai format, err: %v", err))
		}
		// convert to: anthropic format request
		bedrockReq := BedrockRequest{
			MaxTokens:        openaiReq.MaxTokens,
			Temperature:      openaiReq.Temperature,
			TopP:             openaiReq.TopP,
			StopSequences:    openaiReq.Stop,
			AnthropicVersion: defaultBedrockVersion,
		}
		if openaiReq.Temperature <= 0 {
			bedrockReq.Temperature = defaultTemperature
		}
		if openaiReq.MaxTokens <= 1 {
			bedrockReq.MaxTokens = defaultMaxTokens
		}
		// tool
		if len(openaiReq.Tools) > 0 {
			bedrockReq.Tools = openaiReq.Tools
		}
		if openaiReq.ToolChoice != nil {
			bedrockReq.ToolChoice = openaiReq.ToolChoice
		}
		// split system prompt out, keep user / assistant messages
		var systemPrompts []string
		for _, msg := range openaiReq.Messages {
			switch msg.Role {
			case openai.ChatMessageRoleSystem:
				systemPrompts = append(systemPrompts, msg.Content)
			default:
				bedrockMsg, err := common.ConvertOneOpenAIMessage(msg)
				if err != nil {
					panic(fmt.Errorf("failed to convert openai message to bedrock message: %v", err))
				}
				bedrockReq.Messages = append(bedrockReq.Messages, *bedrockMsg)
			}
		}
		if len(systemPrompts) > 0 {
			bedrockReq.System = strings.Join(systemPrompts, "\n")
		}
		anthropicReqBytes, err := json.Marshal(&bedrockReq)
		if err != nil {
			panic(fmt.Errorf("failed to marshal anthropic request: %v", err))
		}
		infor.SetBody2(anthropicReqBytes)

		// get ak/sk
		provider := ctxhelper.MustGetModelProvider(ctx)
		ak := provider.Metadata.Secret["ak"].GetStringValue()
		sk := provider.Metadata.Secret["sk"].GetStringValue()
		if ak == "" || sk == "" {
			panic(fmt.Errorf("missing provider.metadata.secret.{ak,sk}"))
		}
		credCaches := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(ak, sk, ""))
		cred, err := credCaches.Retrieve(ctx)
		if err != nil {
			panic(fmt.Errorf("failed to retrieve aws credentials: %v", err))
		}
		location := provider.Metadata.Public["location"].GetStringValue()
		if location == "" {
			panic(fmt.Errorf("missing provider.metadata.public.location"))
		}

		// payload hash
		var payloadHash string
		sum := sha256.Sum256(anthropicReqBytes)
		payloadHash = hex.EncodeToString(sum[:])
		req.Header.Set("X-Amz-Content-Sha256", payloadHash)

		// remove headers not required for AWS SigV4 signing
		// Keep only known safe headers
		keepHeaders := map[string]bool{
			"host":                 true,
			"content-type":         true,
			"content-length":       true,
			"accept":               true,
			"x-amz-date":           true,
			"x-amz-content-sha256": true,
		}
		for k := range req.Header {
			if !keepHeaders[strings.ToLower(k)] {
				req.Header.Del(k)
			}
		}

		// do aws sign v4
		signer := v4.NewSigner()
		if err := signer.SignHTTP(ctx, cred, req, payloadHash, "bedrock", location, time.Now(),
			func(options *v4.SignerOptions) {
				options.LogSigning = true
				options.Logger = logging.NewStandardLogger(os.Stdout)
			},
		); err != nil {
			panic(fmt.Sprintf("failed to sign request: %v", err))
		}
	})

	return nil
}

func (f *BedrockDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	// non-stream
	if !ctxhelper.GetIsStream(ctx) {
		f.DefaultResponseFilter.Buffer.Write(chunk) // write to buffer, so we can get allChunks later
		return reverseproxy.Continue, nil
	}
	// stream
	var chunkWriter bytes.Buffer
	openaiChunks, err := f.pipeBedrockStream(ctx, io.NopCloser(bytes.NewBuffer(chunk)), &chunkWriter)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to parse bedrock eventstream, err: %v", err)
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

func (f *BedrockDirector) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (err error) {
	// only stream style need append [DONE] chunk
	if !ctxhelper.GetIsStream(ctx) {
		// convert all at once
		var bedrockResp BedrockResponse
		if err := json.Unmarshal(f.DefaultResponseFilter.Buffer.Bytes(), &bedrockResp); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %s, err: %v", string(chunk), err)
		}
		openaiResp, err := bedrockResp.ConvertToOpenAIFormat(ctxhelper.MustGetModel(ctx).Metadata.Public["model_id"].GetStringValue())
		if err != nil {
			return fmt.Errorf("failed to convert anthropic response body to openai format, err: %v", err)
		}
		b, err := json.Marshal(openaiResp)
		if err != nil {
			return fmt.Errorf("failed to marshal openai resp, err: %v", err)
		}
		infor.Header().Del("Content-Length") // remove content-length header, because we will write chunked response
		return f.DefaultResponseFilter.OnResponseEOF(ctx, nil, w, b)
	}
	// append [DONE] chunk
	doneChunk := vars.ConcatChunkDataPrefix([]byte("[DONE]"))
	if _, err := w.Write(doneChunk); err != nil {
		return fmt.Errorf("failed to write openai chunk, err: %v", err)
	}
	return nil
}
