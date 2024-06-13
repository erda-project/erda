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

package excerptor

import (
	"encoding/json"

	"github.com/sashabaranov/go-openai"
)

const (
	ExcerptedImagePlaceholder = "(image content too long, omitted)"
)

func ExcerptActualRequestBody(body string) string {
	var oreq openai.ChatCompletionRequest
	if err := json.Unmarshal([]byte(body), &oreq); err != nil {
		return body
	}
	if len(oreq.Messages) == 0 { // no messages, means json parse failed
		return body
	}
	for i, msg := range oreq.Messages {
		for j, part := range msg.MultiContent {
			if part.ImageURL != nil {
				// it's better to judge url length directly instead of prefix (https://, http://, or other file://)
				if len(part.ImageURL.URL) > 1000 {
					oreq.Messages[i].MultiContent[j].ImageURL.URL = ExcerptedImagePlaceholder
				}
			}
		}
	}
	b, err := json.Marshal(&oreq)
	if err != nil {
		return body
	}
	return string(b)
}
