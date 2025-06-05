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
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/private/protocol/eventstream"
	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// BedrockChunkPayload is AWS Bedrock Claude Streaming chunk.
// see: https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_InvokeModelWithResponseStream.html
type BedrockChunkPayload struct {
	Bytes string `json:"bytes"` // Base64
	// Other fields, like `p`, parse on-demand
}

type BytesRaw struct {
	Type  string            `json:"type"`
	Delta map[string]string `json:"delta"`
	// Other fields, like `index`, see: `raw` at file end.
}

// pipe：Bedrock EventStream →  OpenAI SSE
func (f *BedrockDirector) pipeBedrockStream(ctx context.Context, awsChunkBody io.ReadCloser, w io.Writer) ([]openai.ChatCompletionStreamResponse, error) {
	defer awsChunkBody.Close()

	decoder := eventstream.NewDecoder(bufio.NewReader(awsChunkBody))

	var openaiChunks []openai.ChatCompletionStreamResponse

	for {
		var chunk []byte
		msg, err := decoder.Decode(chunk)
		if err == io.EOF {
			return openaiChunks, nil
		}
		if err != nil {
			return nil, fmt.Errorf("decode eventstream: %w", err)
		}

		et := msg.Headers.Get(":event-type")
		if et == nil {
			return nil, fmt.Errorf("invalid event-type, chunk: %s", string(chunk))
		}
		ctxhelper.GetLogger(ctx).Infof("et: %v", et.String())
		if et.String() != "chunk" {
			continue
		}

		// Claude Messages payload = JSON
		var cp BedrockChunkPayload
		if err := json.Unmarshal(msg.Payload, &cp); err != nil {
			return nil, fmt.Errorf("unmarshal payload: %w", err)
		}

		// see `raw` at file end
		raw, err := base64.StdEncoding.DecodeString(cp.Bytes)
		if err != nil {
			return nil, fmt.Errorf("base64 decode: %w", err)
		}

		var rawObj map[string]any
		if err := json.Unmarshal(raw, &rawObj); err != nil {
			return nil, fmt.Errorf("failed to parse raw, err: %v", err)
		}

		switch rawObj["type"].(string) {
		case "message_start":
			// get message id and model
			f.StreamMessageID = rawObj["message"].(map[string]any)["id"].(string)
			f.StreamMessageModel = rawObj["message"].(map[string]any)["model"].(string)
			f.StreamMessageRole = rawObj["message"].(map[string]any)["role"].(string)
		case "content_block_delta":
			// get delta
			deltaObj := rawObj["delta"].(map[string]any)
			if deltaObj["type"] != "text_delta" {
				continue
			}
			deltaText := deltaObj["text"].(string)
			openaiChunk := openai.ChatCompletionStreamResponse{
				ID:      f.StreamMessageID,
				Object:  "chat.completion.chunk", // fixed
				Created: time.Now().Unix(),
				Model:   f.StreamMessageModel,
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Index: 0,
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Content: deltaText,
							Role:    f.StreamMessageRole,
						},
					},
				},
			}
			openaiChunks = append(openaiChunks, openaiChunk)
		}
	}
}

// raw: {"type":"message_start","message":{"id":"msg_bdrk_01UwU6zwRiDXiYpdxPZSSHyq","type":"message","role":"assistant","model":"claude-3-sonnet-20240229","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":8,"output_tokens":1}}}
// raw: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"I"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"'m"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" ready"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" for"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" your"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" test"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" What"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" woul"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"d you"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" like"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" me"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" to"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" do"}}
// raw: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" ?"}}
// raw: {"type":"content_block_stop","index":0}
// raw: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":28}}
// raw: {"type":"message_stop","amazon-bedrock-invocationMetrics":{"inputTokenCount":8,"outputTokenCount":28,"invocationLatency":968,"firstByteLatency":251}}
