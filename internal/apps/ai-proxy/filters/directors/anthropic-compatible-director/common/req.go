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

package common

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

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
