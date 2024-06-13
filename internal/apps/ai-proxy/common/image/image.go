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

package image

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/assets"
	"github.com/erda-project/erda/pkg/numeral"
)

func HandleChatImage(req *openai.ChatCompletionRequest) error {
	if req == nil {
		return nil
	}
	if !assets.Available() {
		return fmt.Errorf("assets service is not available")
	}
	for i, msg := range req.Messages {
		for j, part := range msg.MultiContent {
			if part.Type != openai.ChatMessagePartTypeImageURL || part.ImageURL == nil {
				continue
			}
			// base64 content has `data:image[/xxx];base64,` prefix, `[]` is optional
			// > "Invalid image URL. The URL must be a valid HTTP or HTTPS URL, or a data URL with base64 encoding."
			s := part.ImageURL.URL
			if !strings.HasPrefix(s, "data:image") {
				continue
			}
			ss := strings.Split(s, "base64,")
			if len(ss) != 2 {
				cutdownNum := int(numeral.MinFloat64([]float64{float64(len(s)), float64(len("data:image/png;base64,") + 10)}))
				return fmt.Errorf("invalid base64 image url content: %s", s[:cutdownNum])
			}
			rawBase64Encoded := ss[1]
			// store image and get download url
			// decode base64 content
			decodedBytes, err := base64.StdEncoding.DecodeString(rawBase64Encoded)
			if err != nil {
				return fmt.Errorf("failed to decode base64 content: %v", err)
			}
			// get download url
			downloadURL, err := assets.Upload("chat-image.png", decodedBytes)
			if err != nil {
				return fmt.Errorf("failed to upload image: %v", err)
			}
			// replace image url
			req.Messages[i].MultiContent[j].ImageURL.URL = downloadURL
		}
	}
	return nil
}
