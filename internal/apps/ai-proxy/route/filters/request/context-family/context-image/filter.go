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
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Name = "context-image"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Context struct{}

var Creator filter_define.RequestRewriterCreator = func(names string, configJSON json.RawMessage) filter_define.ProxyRequestRewriter {
	var cfg Context
	if err := yaml.Unmarshal(configJSON, &cfg); err != nil {
		panic(err)
	}
	return &cfg
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// check request
	bodyCopy, err := body_util.SmartCloneBody(&pr.In.Body, body_util.MaxSample)
	if err != nil {
		return fmt.Errorf("failed to clone request body: %w", err)
	}
	var req openai.ImageRequest
	if err := json.NewDecoder(bodyCopy).Decode(&req); err != nil {
		return err
	}
	// prompt
	if strings.TrimSpace(req.Prompt) == "" {
		return fmt.Errorf("prompt is empty")
	}

	model := ctxhelper.MustGetModel(pr.In.Context())

	// model name
	modelName := model.Name
	if v := model.Metadata.Public["model_name"].GetStringValue(); v != "" {
		modelName = v
	}
	req.Model = modelName

	// response_format
	// - gpt-image-1 not support this param
	if req.Model == "gpt-image-1" {
		req.ResponseFormat = ""
	}

	// reset body
	if err := body_util.SetBody(pr.Out, req); err != nil {
		return fmt.Errorf("failed to set req body for set json body model name, err: %v", err)
	}

	if sink, ok := ctxhelper.GetAuditSink(pr.In.Context()); ok {
		sink.Note("prompt", req.Prompt)
		sink.Note("image.quality", req.Quality)
		sink.Note("image.size", req.Size)
		sink.Note("image.style", req.Style)
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
