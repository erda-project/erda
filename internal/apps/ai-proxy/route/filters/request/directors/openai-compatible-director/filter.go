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

package openai_compatible_director

import (
	"encoding/json"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_segment_getter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/custom-http-director"
)

type OpenaiCompatibleDirector struct {
	*custom_http_director.CustomHTTPDirector
}

var (
	_ filter_define.ProxyRequestRewriter = (*OpenaiCompatibleDirector)(nil)
)

var Creator filter_define.RequestRewriterCreator = func(name string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &OpenaiCompatibleDirector{CustomHTTPDirector: custom_http_director.New()}
}

func init() {
	filter_define.RegisterFilterCreator("openai-compatible-director", Creator)
}

func (f *OpenaiCompatibleDirector) Enable(pr *httputil.ProxyRequest) bool {
	apiSegment := api_segment_getter.GetAPISegment(pr.In.Context())
	return apiSegment != nil &&
		strings.EqualFold(string(apiSegment.APIStyle), string(api_style.APIStyleOpenAICompatible))
}

func (f *OpenaiCompatibleDirector) OnProxyRequest(pr *httputil.ProxyRequest) error {
	if !f.Enable(pr) {
		return nil
	}
	return f.CustomHTTPDirector.OnProxyRequest(pr)
}
