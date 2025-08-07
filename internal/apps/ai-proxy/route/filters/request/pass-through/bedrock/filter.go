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
	"io"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_segment_getter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/aws_bedrock"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/common/xff"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/custom-http-director"
	openai_v1_models "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/openai-v1-models"
)

const (
	Name = "pass-through-bedrock"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Context struct {
	customHTTPDirector *custom_http_director.CustomHTTPDirector
}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{
		customHTTPDirector: custom_http_director.New(),
	}
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// get APIStyle
	apiSegment := api_segment_getter.GetAPISegment(pr.In.Context())
	if apiSegment == nil {
		return fmt.Errorf("api segment not found")
	}
	if !strings.EqualFold(string(apiSegment.APIStyle), string(api_style.APIStyleAnthropicCompatible)) {
		return fmt.Errorf("api style is not AnthropicCompatible")
	}
	// do general non-body fields modify
	if err := f.customHTTPDirector.OnProxyRequest(pr); err != nil {
		return fmt.Errorf("custom http director on error: %w", err)
	}
	// replace path model name
	pathModelName, ok := ctxhelper.GetPathParam(pr.In.Context(), "model")
	if !ok {
		return fmt.Errorf("path variable `model` not found")
	}
	model := ctxhelper.MustGetModel(pr.In.Context())
	modelID := openai_v1_models.GetModelID(model)
	pr.Out.URL.Path = strings.ReplaceAll(pr.Out.URL.Path, pathModelName, modelID)

	bodyBytes, err := io.ReadAll(pr.In.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %v", err)
	}

	// re-sign with provider's credentials
	if err := aws_bedrock.SignAwsRequest(pr, bodyBytes); err != nil {
		return fmt.Errorf("failed to re-sign AWS request: %v", err)
	}

	// force set internal pod ip as xff for anthropic region restriction
	xff.InjectPodIPXFF(pr)

	return nil
}
