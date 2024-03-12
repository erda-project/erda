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

	"github.com/sashabaranov/go-openai"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "context-assistant"
)

var (
	_ reverseproxy.RequestFilter = (*Context)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

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
	// check request: create assistant
	if common.GetRequestRoutePath(ctx) == common.RequestPathPrefixV1Assistants && infor.Method() == http.MethodPost {
		var req openai.AssistantRequest
		if err := json.NewDecoder(infor.BodyBuffer()).Decode(&req); err != nil {
			return reverseproxy.Intercept, err
		}
		// use instructions as prompt
		if req.Instructions != nil && strings.TrimSpace(*req.Instructions) != "" {
			ctxhelper.PutUserPrompt(ctx, *req.Instructions)
		}
		// force set model name through model name
		model, _ := ctxhelper.GetModel(ctx)
		req.Model = model.Name
		infor.SetBody2(req)
	}

	// check request: create thread

	// check request: create run

	// check request: create message
	if common.GetRequestRoutePath(ctx) == common.RequestPathV1ThreadCreateMessage && infor.Method() == http.MethodPost {
		var req openai.MessageRequest
		if err := json.NewDecoder(infor.BodyBuffer()).Decode(&req); err != nil {
			return reverseproxy.Intercept, err
		}
		// use content as prompt
		if strings.TrimSpace(req.Content) != "" {
			ctxhelper.PutUserPrompt(ctx, req.Content)
		}
	}

	return reverseproxy.Continue, nil
}
