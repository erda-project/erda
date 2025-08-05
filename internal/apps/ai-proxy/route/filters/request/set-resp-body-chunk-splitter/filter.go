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

package set_resp_body_chunk_splitter

import (
	"encoding/json"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

type Filter struct {
}

var (
	_ filter_define.ProxyRequestRewriter = (*Filter)(nil)
)

var RequestRewriterCreator filter_define.RequestRewriterCreator = func(name string, _ json.RawMessage) filter_define.ProxyRequestRewriter { return &Filter{} }

func init() {
	filter_define.RegisterFilterCreator("set-response-chunk-splitter", RequestRewriterCreator)
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	splitter := ctxhelper.GetRespBodyChunkSplitter(pr.Out.Context())
	if splitter != nil {
		return nil
	}
	// set default splitter based on streaming or not
	if ctxhelper.GetIsStream(pr.Out.Context()) {
		ctxhelper.PutRespBodyChunkSplitter(pr.Out.Context(), &SSESplitter{})
	} else {
		ctxhelper.PutRespBodyChunkSplitter(pr.Out.Context(), &WholeStreamSplitter{})
	}

	return nil
}
