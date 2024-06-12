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

package dashscope_director

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/image"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/dashscope-director/sdk"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *DashScopeDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)
	requestType := modelMeta.Public.RequestType

	if ok, err := requestType.Valid(); !ok {
		return reverseproxy.Intercept, fmt.Errorf("metadata.public.request_type is invalid, err: %v", err)
	}

	return oneDirector(ctx, w, infor)
}

func oneDirector(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	prov := ctxhelper.MustGetModelProvider(ctx)
	provMeta := getProviderMeta(prov)
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)
	modelName := modelMeta.Public.ModelName

	// rewrite url
	reverseproxy.AppendDirectors(ctx, func(r *http.Request) {
		// rewrite url
		var targetURL string
		if modelMeta.Public.CustomURL != "" {
			targetURL = modelMeta.Public.CustomURL
		} else {
			switch modelMeta.Public.RequestType {
			case metadata.AliyunDashScopeRequestTypeOpenAI:
				targetURL = fmt.Sprintf("%s/compatible-mode/v1/chat/completions", strings.TrimSuffix(provMeta.Public.Endpoint, "/"))
			case metadata.AliyunDashScopeRequestTypeDs:
				var generationType string
				switch model.Type {
				case pb.ModelType_text_generation:
					generationType = "text"
				case pb.ModelType_multimodal:
					generationType = "multimodal"
				default:
					panic(fmt.Sprintf("unsupported dashscope model type: %s", model.Type))
				}
				targetURL = fmt.Sprintf("%s/api/v1/services/aigc/%s-generation/generation", strings.TrimSuffix(provMeta.Public.Endpoint, "/"), generationType)
			}
		}
		u, _ := url.Parse(targetURL)
		r.URL = u
		r.Host = u.Host
		r.Header.Set(httputil.HeaderKeyContentType, string(httputil.ApplicationJsonUTF8))
		r.Header.Set(httputil.HeaderKeyAuthorization, vars.ConcatBearer(prov.ApiKey))
		if ctxhelper.GetIsStream(ctx) {
			r.Header.Set("X-DashScope-SSE", "enable") // stream
		}
		r.Header.Del(httputil.HeaderKeyAcceptEncoding) // disable compression, or we need to handle it (decompress and compress)
	})

	// rewrite body
	var oreq openai.ChatCompletionRequest
	if err := json.NewDecoder(infor.Body()).Decode(&oreq); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to parse request body as openai format, err: %v", err)
	}
	oreq.Model = string(modelName)
	// 需要分析图片，图片只支持 URL，需要把 base64 的内容转换为 URL
	if err := image.HandleChatImage(&oreq); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to handle chat image, err: %v", err)
	}
	var bodyObj any
	switch modelMeta.Public.RequestType {
	case metadata.AliyunDashScopeRequestTypeOpenAI:
		bodyObj = oreq
	case metadata.AliyunDashScopeRequestTypeDs:
		qwreq, err := sdk.ConvertOpenAIChatRequestToDsRequest(oreq, model.Type)
		if err != nil {
			return reverseproxy.Intercept, fmt.Errorf("failed to convert openai chat request to dashscope request, err: %v", err)
		}
		bodyObj = qwreq
	default:
		return reverseproxy.Intercept, fmt.Errorf("unsupported metadata.public.request_type: %s", modelMeta.Public.RequestType)
	}
	b, err := json.Marshal(&bodyObj)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to marshal request body, err: %v", err)
	}
	infor.SetBody(io.NopCloser(bytes.NewBuffer(b)), int64(len(b)))

	return reverseproxy.Continue, nil
}
