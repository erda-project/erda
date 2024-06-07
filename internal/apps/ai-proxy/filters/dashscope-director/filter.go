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

	// 根据模型来定制化处理请求
	// 灵积下模型名唯一，所以可以直接用名字来判断
	if modelMeta.Public.ModelName == metadata.AliyunDashScopeModelNameQwenLong {
		return qwenLongDirector(ctx, w, infor)
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
