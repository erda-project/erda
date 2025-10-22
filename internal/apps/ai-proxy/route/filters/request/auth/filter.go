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

package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

const Name = "auth"

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

type Context struct {
}

var ContextCreator filter_define.RequestRewriterCreator = func(name string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{}
}

func init() {
	filter_define.RegisterFilterCreator(Name, ContextCreator)
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.Out.Context()

	// find client
	clientToken, client, err := akutil.CheckAkOrToken(ctx, pr.In, ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager))
	if err != nil {
		return http_error.NewHTTPError(ctx, http.StatusUnauthorized, err.Error())
	}
	if clientToken != nil {
		ctxhelper.PutClientToken(ctx, clientToken)
	}
	if client == nil {
		return http_error.NewHTTPError(ctx, http.StatusUnauthorized, "client not found")
	} else {
		ctxhelper.PutClient(ctx, client)
		ctxhelper.PutClientId(ctx, client.Id)
		audithelper.NoteOnce(ctx, "client_id", client.Id)
	}

	// user info
	userInfo := getUserInfoFromClientToken(pr)
	for k, v := range userInfo {
		audithelper.Note(ctx, k, v)
	}

	return nil
}

func getUserInfoFromClientToken(pr *httputil.ProxyRequest) map[string]any {
	clientToken, ok := ctxhelper.GetClientToken(pr.Out.Context())
	if !ok || clientToken == nil {
		return nil
	}
	meta := metadata.FromProtobuf(clientToken.Metadata)
	metaCfg := metadata.Config{IgnoreCase: true}
	result := map[string]any{}
	result["dingtalk_staff_id"] = meta.MustGetValueByKey(vars.XAIProxyDingTalkStaffID, metaCfg)
	result["email"] = meta.MustGetValueByKey(vars.XAIProxyEmail, metaCfg)
	result["identity_job_number"] = meta.MustGetValueByKey(vars.XAIProxyJobNumber, metaCfg)
	result["username"] = meta.MustGetValueByKey(vars.XAIProxyName, metaCfg)
	result["identity_phone_number"] = meta.MustGetValueByKey(vars.XAIProxyPhone, metaCfg)
	return result
}
