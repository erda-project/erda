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

package thinking_handler

import (
	"encoding/json"
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	encodersregistry "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders/registry"
	extractorsregistry "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/extractors/registry"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Name = "thinking-handler"
)

var (
	_ filter_define.ProxyRequestRewriter = (*ThinkingHandler)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type ThinkingHandler struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &ThinkingHandler{}
}

var extractorsRegistry = extractorsregistry.NewRegistry()
var encodersRegistry = encodersregistry.NewRegistry()

func (f *ThinkingHandler) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// only process JSON requests
	contentType := pr.Out.Header.Get(httperrorutil.HeaderKeyContentType)
	if !strings.HasPrefix(contentType, string(httperrorutil.ApplicationJson)) {
		return nil
	}

	return f.handleJSONRequest(pr)
}

func (f *ThinkingHandler) handleJSONRequest(pr *httputil.ProxyRequest) error {
	bodyCopy, err := body_util.SmartCloneBody(&pr.Out.Body, body_util.MaxSample)
	if err != nil {
		return fmt.Errorf("failed to clone request body: %w", err)
	}

	var jsonBody map[string]any
	if err := json.NewDecoder(bodyCopy).Decode(&jsonBody); err != nil {
		return fmt.Errorf("failed to decode request body: %v", err)
	}

	// extract thinking configs
	originalFields, commonThinking, err := extractorsRegistry.ExtractAll(jsonBody)
	if err != nil {
		return err
	}
	if commonThinking == nil {
		return nil
	}

	// encode to target format
	appendBodyMap, err := encodersRegistry.EncodeAll(pr.Out.Context(), *commonThinking)
	if err != nil {
		return fmt.Errorf("failed to encode thinking: %v", err)
	}

	// clean original thinking configs
	for originalKey := range originalFields {
		delete(jsonBody, originalKey)
	}
	// append body map
	for k, v := range appendBodyMap {
		jsonBody[k] = v
	}

	// add response header
	ctxhelper.PutRequestThinkingTransformChanges(pr.Out.Context(), map[string]any{
		"from": originalFields,
		"to":   appendBodyMap,
	})

	// marshal and set updated body
	updatedBody, err := json.Marshal(jsonBody)
	if err != nil {
		return fmt.Errorf("failed to marshal updated body: %v", err)
	}

	if err := body_util.SetBody(pr.Out, updatedBody); err != nil {
		return fmt.Errorf("failed to set updated body: %w", err)
	}

	return nil
}
