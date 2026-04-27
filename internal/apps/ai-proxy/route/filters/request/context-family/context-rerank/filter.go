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

package context_rerank

import (
	"encoding/json"
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Name = "context-rerank"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Context struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{}
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	bodyCopy, err := body_util.SmartCloneBody(&pr.In.Body, body_util.MaxSample)
	if err != nil {
		return fmt.Errorf("failed to clone request body: %w", err)
	}
	if bodyCopy.Size() > 0 {
		var req map[string]any
		if err := json.NewDecoder(bodyCopy).Decode(&req); err != nil {
			return fmt.Errorf("failed to decode rerank request body: %w", err)
		}
		if query := stringifyQuery(req["query"]); strings.TrimSpace(query) != "" {
			audithelper.Note(pr.In.Context(), "prompt", query)
		}
	}

	if err := f.trySetJSONBodyModelName(pr); err != nil {
		return fmt.Errorf("failed to set model name in JSON body: %w", err)
	}

	return nil
}

func stringifyQuery(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func (f *Context) trySetJSONBodyModelName(pr *httputil.ProxyRequest) error {
	if !strings.HasPrefix(pr.Out.Header.Get(httperrorutil.HeaderKeyContentType), string(httperrorutil.ApplicationJson)) {
		return nil
	}

	var reqBody map[string]any
	if err := json.NewDecoder(pr.Out.Body).Decode(&reqBody); err != nil {
		l := ctxhelper.MustGetLogger(pr.Out.Context())
		l.Errorf("failed to decode req body for set json body model name")
		return nil
	}

	model := ctxhelper.MustGetModel(pr.Out.Context())
	modelName := model.Name
	if customModelName := model.Metadata.Public["model_name"]; customModelName != nil {
		if v := customModelName.GetStringValue(); v != "" {
			modelName = v
		}
	}
	reqBody["model"] = modelName
	if err := body_util.SetBody(pr.Out, reqBody); err != nil {
		return fmt.Errorf("failed to set req body for set json body model name, err: %w", err)
	}
	return nil
}
