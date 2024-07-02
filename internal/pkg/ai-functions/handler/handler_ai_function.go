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

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	aitestcase "github.com/erda-project/erda/internal/pkg/ai-functions/functions/test-case"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	_ pb.AiFunctionServer = (*AIFunction)(nil)
)

type AIFunction struct {
	Log       logs.Logger
	OpenaiURL *url.URL
	ModelIds  map[string]string
}

const (
	EnvAiProxyChatgpt4ModelId = "AI_PROXY_CHATGPT4_MODEL_ID"
)

func (h *AIFunction) Apply(ctx context.Context, req *pb.ApplyRequest) (pbValue *structpb.Value, err error) {
	h.Log.Infof("apply the function with request %s", strutil.TryGetJsonStr(req))
	var results any

	factory, ok := functions.Retrieve(req.GetFunctionName())
	if !ok {
		err := errors.Errorf("AI function %s not found", req.GetFunctionName())
		return nil, HTTPError(err, http.StatusBadRequest)
	}

	f := factory(ctx, "", req.GetBackground())
	xAIProxyModelId := ""
	switch req.GetFunctionName() {
	case aitestcase.Name:
		// 需要 ChatGPT4 模型进行需求分组
		xAIProxyModelId, _ = h.ModelIds["gpt-4"]
		if xAIProxyModelId == "" {
			h.Log.Errorf("config modelIds[\"gpt-4\"] not set, please set env %s", EnvAiProxyChatgpt4ModelId)
			err = errors.Errorf("config modelIds[\"gpt-4\"] not set")
			return nil, HTTPError(err, http.StatusInternalServerError)
		}
	default:
		err := errors.Errorf("AI function %s not support for apply", req.GetFunctionName())
		return nil, HTTPError(err, http.StatusBadRequest)
	}

	results, err = f.Handler(ctx, factory, req, h.OpenaiURL, xAIProxyModelId)
	if err != nil {
		return nil, HTTPError(err, http.StatusInternalServerError)
	}

	var v = &structpb.Value{}
	marshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	switch i := results.(type) {
	case string:
		if err := marshaler.Unmarshal([]byte(i), v); err != nil {
			return nil, err
		}
	case []byte:
		if err := marshaler.Unmarshal(i, v); err != nil {
			return nil, err
		}
	case json.RawMessage:
		if err := marshaler.Unmarshal(i, v); err != nil {
			return nil, err
		}
	default:
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(results); err != nil {
			return nil, err
		}
		if err := marshaler.Unmarshal(buf.Bytes(), v); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (h *AIFunction) GetSystemPrompt(ctx context.Context, req *pb.GetSystemPromptRequest) (pbValue *structpb.Value, err error) {
	h.Log.Infof("get system prompt for function name %s", req.GetFunctionName())

	factory, ok := functions.Retrieve(req.GetFunctionName())
	if !ok {
		err := errors.Errorf("AI function %s not found", req.GetFunctionName())
		return nil, HTTPError(err, http.StatusBadRequest)
	}
	f := factory(ctx, "", nil)

	content := httpserver.Resp{
		Success: true,
		Data:    f.SystemMessage(apis.GetLang(ctx)), // 系统提示语,
	}
	result, err := json.Marshal(content)
	if err != nil {
		return nil, HTTPError(err, http.StatusInternalServerError)
	}

	var v = &structpb.Value{}
	marshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := marshaler.Unmarshal(result, v); err != nil {
		return nil, err
	}
	return v, nil
}

// HTTPError todo: duplicate
func HTTPError(err error, code int) error {
	if err == nil {
		err = errors.New(http.StatusText(code))
	}
	return httpError{error: err, code: code}
}

type httpError struct {
	error
	code int
}

func (e httpError) HTTPStatus() int {
	return e.code
}
