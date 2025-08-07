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

package checker

import (
	"encoding/json"
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

const name = "check-pass-through"

var (
	_ filter_define.ProxyRequestRewriter = (*Checker)(nil)
)

type Checker struct {
}

func init() {
	filter_define.RegisterFilterCreator(name, Creator)
}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Checker{}
}

func (f *Checker) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// only support auth by client-token, reject client-ak
	token := vars.TrimBearer(pr.Out.Header.Get("Authorization"))
	if !strings.HasPrefix(token, client_token.TokenPrefix) {
		return fmt.Errorf("pass-through usage only support auth by token")
	}
	return nil
}
