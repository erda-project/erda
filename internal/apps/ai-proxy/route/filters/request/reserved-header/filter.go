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

package audit_save_out

import (
	"encoding/json"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

const (
	Name = "reserved-header"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Filter)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Filter struct {
	firstResponseAt time.Time
}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Filter{}
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// del all X-AI-Proxy-* headers before invoking llm
	for k := range pr.Out.Header {
		if strings.HasPrefix(strings.ToUpper(k), strings.ToUpper(vars.XAIProxyHeaderPrefix)) {
			pr.Out.Header.Del(k)
		}
	}
	return nil
}
