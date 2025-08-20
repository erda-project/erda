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
	"net/http/httputil"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "context-embedding"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Context struct {
}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{}
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	bodyCopy, err := body_util.SmartCloneBody(&pr.In.Body, body_util.MaxSample)
	if err != nil {
		return fmt.Errorf("failed to clone request body: %w", err)
	}
	var req openai.EmbeddingRequest
	if err := json.NewDecoder(bodyCopy).Decode(&req); err != nil {
		return fmt.Errorf("failed to decode embedding request, err: %v", err)
	}
	var prompt string
	switch req.Input.(type) {
	case string:
		prompt = req.Input.(string)
	case []interface{}:
		var ss []string
		for _, item := range req.Input.([]interface{}) {
			ss = append(ss, strutil.String(item))
		}
		prompt = strutil.Join(ss, "\n")
	}
	if prompt != "" {
		if sink, ok := ctxhelper.GetAuditSink(pr.In.Context()); ok {
			sink.Note("prompt", prompt)
		}
	}

	// set model name in JSON body
	if err := f.trySetJSONBodyModelName(pr); err != nil {
		return fmt.Errorf("failed to set model name in JSON body: %v", err)
	}

	return nil
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
	var modelName any = model.Name
	if customModelName := model.Metadata.Public["model_name"]; customModelName != nil {
		modelName = customModelName
	}
	reqBody["model"] = modelName
	if err := body_util.SetBody(pr.Out, reqBody); err != nil {
		return fmt.Errorf("failed to set req body for set json body model name, err: %v", err)
	}
	return nil
}
