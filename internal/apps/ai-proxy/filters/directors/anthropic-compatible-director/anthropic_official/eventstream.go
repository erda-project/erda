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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

// pipe：Anthropic EventStream →  OpenAI SSE
func (f *AnthropicDirector) pipeAnthropicStream(ctx context.Context, chunkBody io.ReadCloser, w io.Writer) ([]openai.ChatCompletionStreamResponse, error) {
	defer chunkBody.Close()

	var openaiChunks []openai.ChatCompletionStreamResponse

	// read by line
	scanner := bufio.NewScanner(chunkBody)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		dataLine := string(vars.TrimChunkDataPrefix([]byte(line)))

		var rawObj map[string]any
		if err := json.Unmarshal([]byte(dataLine), &rawObj); err != nil {
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

	return openaiChunks, nil
}

// event: message_start
// data: {"type":"message_start","message":{"id":"msg_015vNuMABCWdvpWgxFLFxPiA","type":"message","role":"assistant","model":"claude-3-haiku-20240307","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":8,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":3,"service_tier":"standard"}}            }
//
// event: content_block_start
// data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}             }
//
// event: ping
// data: {"type": "ping"}
//
// event: content_block_delta
// data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello! How"}         }
//
// event: content_block_delta
// data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" can I assist you today?"}             }
//
// event: content_block_stop
// data: {"type":"content_block_stop","index":0   }
//
// event: message_delta
// data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":12}            }
//
// event: message_stop
// data: {"type":"message_stop"          }
