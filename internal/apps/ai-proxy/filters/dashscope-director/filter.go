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

	"github.com/sashabaranov/go-openai"

	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "dashscope-director"
)

var (
	_ reverseproxy.RequestFilter = (*DashScopeDirector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type DashScopeDirector struct {
	*reverseproxy.DefaultResponseFilter

	lastCompletionDataLineIndex int
	lastCompletionDataLineText  string
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	return &DashScopeDirector{
		DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(),
	}, nil
}

func (f *DashScopeDirector) MultiResponseWriter(ctx context.Context) []io.ReadWriter {
	return []io.ReadWriter{ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)}
}

func (f *DashScopeDirector) Enable(ctx context.Context, req *http.Request) bool {
	prov, ok := ctxhelper.GetModelProvider(ctx)
	return ok && prov.Type == modelproviderpb.ModelProviderType_AliyunDashScope
}

func (f *DashScopeDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)

	// Customize request processing based on the model
	// The names of the models are unique, so they can be directly used to determine.
	switch modelMeta.Public.ModelName {
	case metadata.AliyunDashScopeModelNameQwenLong:
		return qwenLongDirector(ctx, w, infor)
	case metadata.AliyunDashScopeModelNameQwenVLPlus, metadata.AliyunDashScopeModelNameQwenVLMax:
		return qwenVLDirector(ctx, w, infor)
	}

	return reverseproxy.Continue, nil
}

func qwenLongDirector(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	prov := ctxhelper.MustGetModelProvider(ctx)
	provMeta := getProviderMeta(prov)

	// rewrite url
	reverseproxy.AppendDirectors(ctx, func(r *http.Request) {
		// rewrite url
		dashScopeURL := fmt.Sprintf("%s/compatible-mode/v1/chat/completions", strings.TrimSuffix(provMeta.Public.Endpoint, "/"))
		u, _ := url.Parse(dashScopeURL)
		r.URL = u
		r.Host = u.Host
		// rewrite authorization header
		r.Header.Set(httputil.HeaderKeyContentType, string(httputil.ApplicationJsonUTF8))
		r.Header.Set(httputil.HeaderKeyAuthorization, vars.ConcatBearer(prov.ApiKey))
		//r.Header.Set(httputil.HeaderKeyAccept, string(httputil.ApplicationJsonUTF8))
		//r.Header.Del(httputil.HeaderKeyAcceptEncoding) // remove gzip. Actual test: gzip is not ok; deflate is ok; br is ok
		r.Header.Add(httputil.HeaderKeyAcceptEncoding, "gzip")
		r.Header.Add(httputil.HeaderKeyAcceptEncoding, "deflate")
	})

	// rewrite body
	// same as OpenAI format, but we need override model name
	var req map[string]interface{}
	if err := json.NewDecoder(infor.Body()).Decode(&req); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to parse request body as openai format, err: %v", err)
	}
	req["model"] = metadata.AliyunDashScopeModelNameQwenLong
	b, err := json.Marshal(req)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to marshal request body, err: %v", err)
	}
	infor.SetBody(io.NopCloser(bytes.NewBuffer(b)), int64(len(b)))

	return reverseproxy.Continue, nil
}

func qwenVLDirector(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	prov := ctxhelper.MustGetModelProvider(ctx)
	provMeta := getProviderMeta(prov)
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)

	// rewrite url
	reverseproxy.AppendDirectors(ctx, func(r *http.Request) {
		// rewrite url
		dashScopeURL := fmt.Sprintf("%s/api/v1/services/aigc/multimodal-generation/generation", strings.TrimSuffix(provMeta.Public.Endpoint, "/"))
		u, _ := url.Parse(dashScopeURL)
		r.URL = u
		r.Host = u.Host
		// rewrite authorization header
		r.Header.Set(httputil.HeaderKeyContentType, string(httputil.ApplicationJsonUTF8))
		r.Header.Set(httputil.HeaderKeyAuthorization, vars.ConcatBearer(prov.ApiKey))
		if ctxhelper.GetIsStream(ctx) {
			r.Header.Set("X-DashScope-SSE", "enable") // stream
		}
		r.Header.Del(httputil.HeaderKeyAcceptEncoding) // disable compression, or we need to handle it (decompress and compress)
		//r.Header.Set(httputil.HeaderKeyAccept, string(httputil.ApplicationJsonUTF8))
		//r.Header.Del(httputil.HeaderKeyAcceptEncoding) // remove gzip. Actual test: gzip is not ok; deflate is ok; br is ok
		//r.Header.Add(httputil.HeaderKeyAcceptEncoding, "gzip")
		//r.Header.Add(httputil.HeaderKeyAcceptEncoding, "deflate")
	})

	// rewrite body
	var oreq openai.ChatCompletionRequest
	if err := json.NewDecoder(infor.Body()).Decode(&oreq); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to parse request body as openai format, err: %v", err)
	}
	oreq.Model = string(modelMeta.Public.ModelName)
	// 需要分析图片，图片只支持 URL，需要把 base64 的内容转换为 URL
	if err := handleChatImage(&oreq); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to handle chat image, err: %v", err)
	}
	qwreq := convertOpenAIChatRequestToQwenVLRequest(oreq)
	b, err := json.Marshal(&qwreq)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to marshal request body, err: %v", err)
	}
	infor.SetBody(io.NopCloser(bytes.NewBuffer(b)), int64(len(b)))

	return reverseproxy.Continue, nil
}

type (
	QwenVLRequest struct {
		Model      string                  `json:"model,omitempty"`
		Input      QwenVLRequestInput      `json:"input,omitempty"`
		Parameters QwenVLRequestParameters `json:"parameters,omitempty"`
	}
	QwenVLRequestInput struct {
		Messages []QwenVLRequestMessage `json:"messages,omitempty"`
	}
	QwenVLRequestMessage struct {
		Role    string                     `json:"role,omitempty"`
		Content []QwenVLRequestContentItem `json:"content,omitempty"`
	}
	QwenVLRequestContentItem struct {
		Image string `json:"image,omitempty"` // must be URL now
		Text  string `json:"text,omitempty"`
	}
	QwenVLRequestParameters struct {
	}
)

func convertOpenAIChatRequestToQwenVLRequest(oreq openai.ChatCompletionRequest) QwenVLRequest {
	var req QwenVLRequest
	req.Model = oreq.Model
	for _, om := range oreq.Messages {
		var m QwenVLRequestMessage
		m.Role = om.Role
		for _, omc := range om.MultiContent {
			switch omc.Type {
			case openai.ChatMessagePartTypeText:
				m.Content = append(m.Content, QwenVLRequestContentItem{Text: omc.Text})
			case openai.ChatMessagePartTypeImageURL:
				// TODO handle URL here, not support base64
				m.Content = append(m.Content, QwenVLRequestContentItem{Image: omc.ImageURL.URL})
			}
		}
		if len(om.Content) > 0 {
			m.Content = append(m.Content, QwenVLRequestContentItem{Text: om.Content})
		}
		req.Input.Messages = append(req.Input.Messages, m)
	}
	return req
}
