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

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/custom-http-director"
)

const (
	Name = "pass-through-bailian"
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
	if err := f.customHTTPDirector.OnProxyRequest(pr); err != nil {
		return fmt.Errorf("custom http director on error: %w", err)
	}
	return nil
}
