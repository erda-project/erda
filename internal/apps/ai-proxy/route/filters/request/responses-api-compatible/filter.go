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

package responses_api_compatible

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types/common_types_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

const (
	Name = "responses-api-compatible"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Filter)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Filter struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Filter{}
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// handle volcengine-ark
	provider := ctxhelper.MustGetModelProvider(pr.In.Context())
	serviceProviderType := common_types_util.GetServiceProviderType(provider)
	if serviceProviderType == common_types.ServiceProviderTypeVolcengineArk.String() {
		bodyCopy, err := body_util.SmartCloneBody(&pr.Out.Body, body_util.MaxSample)
		if err != nil {
			return fmt.Errorf("failed to clone request body: %w", err)
		}
		defer bodyCopy.Close()

		if bodyCopy.Size() == 0 {
			return nil
		}

		var req map[string]any
		if err := json.NewDecoder(bodyCopy).Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("failed to decode request body: %w", err)
		}
		if req == nil {
			return nil
		}

		needRewrite := false

		if _, hasInclude := req["include"]; hasInclude {
			delete(req, "include")
			needRewrite = true
		}

		if input, ok := req["input"]; ok && input != nil {
			if normalizedInput, changed := normalizeInputMessages(input); changed {
				req["input"] = normalizedInput
				needRewrite = true
			}
		}

		if !needRewrite {
			return nil
		}

		if err := body_util.SetBody(pr.Out, req); err != nil {
			return fmt.Errorf("failed to set request body: %w", err)
		}
	}

	return nil
}

const (
	userRole                  = "user"
	inputContentTypeInputText = "input_text"
)

func normalizeInputMessages(input any) (any, bool) {
	switch v := input.(type) {
	case []any:
		changed := false
		for idx, item := range v {
			normalizedItem, itemChanged := normalizeInputMessage(item)
			if itemChanged {
				changed = true
			}
			v[idx] = normalizedItem
		}
		return v, changed
	case map[string]any:
		normalizedItem, _ := normalizeInputMessage(v)
		return []any{normalizedItem}, true
	case string:
		return []any{createUserMessageFromText(v)}, true
	default:
		return input, false
	}
}

func normalizeInputMessage(item any) (any, bool) {
	switch msg := item.(type) {
	case map[string]any:
		changed := false

		if _, has := msg["id"]; has {
			delete(msg, "id")
			changed = true
		}
		if _, has := msg["annotations"]; has {
			delete(msg, "annotations")
			changed = true
		}
		if _, has := msg["logprobs"]; has {
			delete(msg, "logprobs")
			changed = true
		}

		if typ, ok := msg["type"].(string); ok && typ != "" && !strings.EqualFold(typ, "message") {
			return msg, changed
		}

		if role, ok := msg["role"].(string); !ok || !strings.EqualFold(role, userRole) {
			msg["role"] = userRole
			changed = true
		} else if role != userRole {
			msg["role"] = userRole
			changed = true
		}

		var (
			normalizedContent any
			contentChanged    bool
		)
		if content, exists := msg["content"]; exists {
			normalizedContent, contentChanged = normalizeInputContent(content)
		} else {
			normalizedContent, contentChanged = normalizeInputContent(nil)
		}
		if contentChanged {
			changed = true
		}
		msg["content"] = normalizedContent

		return msg, changed
	case string:
		return createUserMessageFromText(msg), true
	case []any:
		content, _ := normalizeInputContent(msg)
		return map[string]any{
			"role":    userRole,
			"content": content,
		}, true
	default:
		return item, false
	}
}

func normalizeInputContent(content any) (any, bool) {
	switch c := content.(type) {
	case string:
		return []any{createContentPartFromText(c)}, true
	case []any:
		changed := false
		for idx, part := range c {
			normalizedPart, partChanged := normalizeContentPart(part)
			if partChanged {
				changed = true
			}
			c[idx] = normalizedPart
		}
		return c, changed
	case map[string]any:
		normalizedPart, _ := normalizeContentPart(c)
		return []any{normalizedPart}, true
	case nil:
		return []any{createContentPartFromText("")}, true
	default:
		return []any{createContentPartFromText(fmt.Sprint(c))}, true
	}
}

func createUserMessageFromText(text string) map[string]any {
	return map[string]any{
		"role": userRole,
		"content": []any{
			createContentPartFromText(text),
		},
	}
}

func createContentPartFromText(text string) map[string]any {
	return map[string]any{
		"type": inputContentTypeInputText,
		"text": text,
	}
}

func normalizeContentPart(part any) (map[string]any, bool) {
	switch pv := part.(type) {
	case map[string]any:
		changed := false
		if typ, ok := pv["type"].(string); !ok || !strings.EqualFold(typ, inputContentTypeInputText) {
			pv["type"] = inputContentTypeInputText
			changed = true
		} else if typ != inputContentTypeInputText {
			pv["type"] = inputContentTypeInputText
			changed = true
		}
		if _, has := pv["annotations"]; has {
			delete(pv, "annotations")
			changed = true
		}
		if _, has := pv["logprobs"]; has {
			delete(pv, "logprobs")
			changed = true
		}
		if _, ok := pv["text"].(string); !ok {
			textValue := fmt.Sprint(pv["text"])
			if pv["text"] == nil || textValue == "<nil>" {
				textValue = ""
			}
			pv["text"] = textValue
			changed = true
		}
		return pv, changed
	case string:
		return createContentPartFromText(pv), true
	default:
		return createContentPartFromText(fmt.Sprint(pv)), true
	}
}
