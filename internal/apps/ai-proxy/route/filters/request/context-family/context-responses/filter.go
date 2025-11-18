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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

// ContentType represents the type of content in the message
type ContentType string

const (
	Name                             = "context-responses"
	ContentTypeInputText ContentType = "input_text"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
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

type Context struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{}
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	if ctxhelper.MustGetPathMatcher(pr.In.Context()).Pattern == vars.RequestPathPrefixV1Responses && pr.In.Method == http.MethodPost {
		bodyCopy, err := body_util.SmartCloneBody(&pr.In.Body, body_util.MaxSample)
		if err != nil {
			return fmt.Errorf("failed to clone request body: %w", err)
		}
		if bodyCopy.Size() == 0 {
			return fmt.Errorf("request body is empty")
		}
		var req map[string]any
		if err := json.NewDecoder(bodyCopy).Decode(&req); err != nil {
			return err
		}

		prompts := make([]string, 0)

		if instructions, ok := req["instructions"].(string); ok && strings.TrimSpace(instructions) != "" {
			prompts = append(prompts, instructions)
		}

		// Handle OpenAI Responses API input structure
		if input, ok := req["input"]; ok {
			prompts = append(prompts, FindUserPrompts(input)...)
		}

		isStream := false
		if v := req["stream"]; v != nil && v.(bool) {
			isStream = v.(bool)
		}
		ctxhelper.PutIsStream(pr.Out.Context(), isStream)

		// set model name in JSON body
		if err := f.trySetJSONBodyModelName(pr); err != nil {
			return fmt.Errorf("failed to set model name in JSON body: %v", err)
		}

		// prompt
		prompt := strings.Join(prompts, "\n")
		audithelper.Note(pr.In.Context(), "prompt", prompt)
	}

	return nil
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

func (f *Context) trySetJSONBodyModelName(pr *httputil.ProxyRequest) error {
	if !strings.HasPrefix(pr.Out.Header.Get(httperrorutil.HeaderKeyContentType), string(httperrorutil.ApplicationJson)) {
		return nil
	}
	// update model name
	var reqBody map[string]any
	if err := json.NewDecoder(pr.Out.Body).Decode(&reqBody); err != nil {
		l := ctxhelper.MustGetLogger(pr.Out.Context())
		l.Errorf("failed to decode req body for set json body model name")
		return nil
	}
	model := ctxhelper.MustGetModel(pr.Out.Context())
	modelName := model.Name
	if customModelName := model.Metadata.Public["model_name"]; customModelName != nil {
		modelName = customModelName.GetStringValue()
	}
	reqBody["model"] = modelName
	if err := body_util.SetBody(pr.Out, reqBody); err != nil {
		return fmt.Errorf("failed to set req body for set json body model name, err: %v", err)
	}
	return nil
}
