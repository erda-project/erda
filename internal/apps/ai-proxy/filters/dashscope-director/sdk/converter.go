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

package sdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func ConvertOpenAIChatRequestToDsRequest(ctx context.Context, oreq openai.ChatCompletionRequest, targetModelType pb.ModelType) (DsRequest, error) {
	var dsReq DsRequest
	dsReq.Model = oreq.Model
	for _, om := range oreq.Messages {
		var m DsRequestMessage
		m.Role = om.Role
		switch targetModelType {
		case pb.ModelType_text_generation:
			var content string
			// string content
			if om.Content != "" {
				content = om.Content
			} else {
				// multi part content
				var textParts []string
				for _, omc := range om.MultiContent {
					if omc.Text != "" {
						textParts = append(textParts, omc.Text)
					}
				}
				content = strings.Join(textParts, "\n")
			}
			m.Content = content

		case pb.ModelType_multimodal:
			var parts []DsRequestContentPart
			// string content
			if om.Content != "" {
				part := DsRequestContentPart{Text: om.Content}
				parts = append(parts, part)
			} else {
				// multi part content
				for _, omc := range om.MultiContent {
					switch omc.Type {
					case openai.ChatMessagePartTypeText:
						parts = append(parts, DsRequestContentPart{Text: omc.Text})
					case openai.ChatMessagePartTypeImageURL:
						parts = append(parts, DsRequestContentPart{Image: omc.ImageURL.URL})
					default:
						ctxhelper.GetLogger(ctx).Warnf("unsupported message part type: %s", omc.Type)
					}
				}
			}
			m.Content = parts

		default:
			return DsRequest{}, fmt.Errorf("unsupported target model type: %s", targetModelType)
		}
		dsReq.Input.Messages = append(dsReq.Input.Messages, m)
	}
	return dsReq, nil
}

func ConvertDsStreamChunkToOpenAIFormat(dsChunk DsRespStreamChunk, modelName string) (*openai.ChatCompletionStreamResponse, error) {
	var ocs []openai.ChatCompletionStreamChoice
	for _, dsc := range dsChunk.Output.Choices {
		oc := openai.ChatCompletionStreamChoice{
			Index: 0,
			Delta: openai.ChatCompletionStreamChoiceDelta{
				Content: dsc.Message.Content.(string),
				Role:    dsc.Message.Role,
			},
			FinishReason: openai.FinishReason(dsc.FinishReason),
		}
		ocs = append(ocs, oc)
	}
	openaiChunk := openai.ChatCompletionStreamResponse{
		ID:      dsChunk.RequestID,
		Object:  "chat.completion.chunk", // fixed
		Model:   modelName,
		Choices: ocs,
		Usage: &openai.Usage{
			PromptTokens:     int(dsChunk.Usage.InputTokens),
			CompletionTokens: int(dsChunk.Usage.OutputTokens),
			TotalTokens:      int(dsChunk.Usage.TotalTokens),
		},
	}
	return &openaiChunk, nil
}
