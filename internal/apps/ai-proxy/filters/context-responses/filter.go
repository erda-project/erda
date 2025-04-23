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

package context

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

// ContentType represents the type of content in the message
type ContentType string

const (
	Name                             = "context-responses"
	ContentTypeInputText ContentType = "input_text"
)

var (
	_ reverseproxy.RequestFilter = (*Context)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type (
	// Message represents a message in the conversation
	Message struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	}
	// ContentObject represents a content object in the message
	ContentObject struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
)

type Context struct {
	Config *Config
}

type Config struct {
}

func New(configJSON json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(configJSON, &cfg); err != nil {
		return nil, err
	}
	return &Context{Config: &cfg}, nil
}

func (f *Context) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	if common.GetRequestRoutePath(ctx) == common.RequestPathPrefixV1Responses && infor.Method() == http.MethodPost {
		var req map[string]any
		if err := json.NewDecoder(infor.BodyBuffer()).Decode(&req); err != nil {
			return reverseproxy.Intercept, err
		}

		prompts := make([]string, 0)

		if instructions, ok := req["instructions"].(string); ok && strings.TrimSpace(instructions) != "" {
			prompts = append(prompts, instructions)
		}

		// Handle OpenAI Responses API input structure
		if input, ok := req["input"]; ok {
			prompts = append(prompts, FindUserPrompts(input)...)
		}

		ctxhelper.PutUserPrompt(ctx, strings.Join(prompts, "\n"))
		infor.SetBody2(req)
	}

	return reverseproxy.Continue, nil
}

// FindUserPrompts recursively extracts user prompts from OpenAI Responses API structure
func FindUserPrompts(obj any) []string {
	if obj == nil {
		return nil
	}

	prompts := make([]string, 0)
	switch v := obj.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			prompts = append(prompts, v)
		}
	case []any:
		for _, item := range v {
			if msg, ok := item.(map[string]any); ok {
				if role, ok := msg["role"].(string); ok && role == "user" {
					if content, ok := msg["content"]; ok {
						prompts = append(prompts, extractPromptsFromContent(content)...)
					}
				}
			}
		}
	}

	return prompts
}

// extractPromptsFromContent extracts prompts from content of different types
func extractPromptsFromContent(content any) []string {
	prompts := make([]string, 0)

	switch c := content.(type) {
	case string:
		if strings.TrimSpace(c) != "" {
			prompts = append(prompts, c)
		}
	case []any:
		for _, cc := range c {
			if contentObj, ok := cc.(map[string]any); ok {
				if contentType, ok := contentObj["type"].(string); ok && ContentType(contentType) == ContentTypeInputText {
					if text, ok := contentObj["text"].(string); ok && text != "" {
						prompts = append(prompts, text)
					}
				}
			}
		}
	}

	return prompts
}
