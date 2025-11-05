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
	"net/http"
	"net/http/httputil"

	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

type Context struct {
}

var ContextCreator filter_define.RequestRewriterCreator = func(name string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{}
}

func init() {
	filter_define.RegisterFilterCreator("context", ContextCreator)
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.Out.Context()

	var (
		l = ctxhelper.MustGetLogger(ctx)
		q = ctxhelper.MustGetDBClient(ctx)
	)

	// session
	var session *sessionpb.Session
	headerSessionId := pr.Out.Header.Get(vars.XAIProxySessionId)
	if headerSessionId != "" && headerSessionId != vars.UIValueUndefined {
		_session, err := q.SessionClient().Get(ctx, &sessionpb.SessionGetRequest{Id: headerSessionId})
		if err != nil {
			l.Errorf("failed to get session, id: %s, err: %v", headerSessionId, err)
			return http_error.NewHTTPError(ctx, http.StatusBadRequest, "SessionId is invalid")
		}
		session = _session
	}

	// find model
	model, err := findModel(pr.In, pr.In.Context(), ctxhelper.MustGetClient(ctx))
	if err != nil {
		l.Errorf("failed to request model, err: %v", err)
		return http_error.NewHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("Model is invalid: %v", err))
	}
	model, err = cachehelpers.GetRenderedModelByID(ctx, model.Id)
	if err != nil {
		l.Errorf("failed to get rendered model, err: %v", err)
		return http_error.NewHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("ModelId is invalid: %v", err))
	}

	// find provider
	serviceProvider, err := cachehelpers.GetRenderedServiceProviderByID(ctx, model.ProviderId)
	if err != nil {
		l.Errorf("failed to get service provider, id: %s, err: %v", model.ProviderId, err)
		return http_error.NewHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("ServiceProviderId is invalid: %v", err))
	}

	// find prompt
	headerPromptId := pr.Out.Header.Get(vars.XAIProxyPromptId)
	var prompt *promptpb.Prompt
	if headerPromptId != "" {
		_prompt, err := q.PromptClient().Get(ctx, &promptpb.PromptGetRequest{Id: headerPromptId})
		if err != nil {
			l.Errorf("failed to get prompt, id: %s, err: %v", headerPromptId, err)
			return http_error.NewHTTPError(ctx, http.StatusBadRequest, "Prompt is invalid")
		}
		prompt = _prompt
	}

	// store data to context
	ctxhelper.PutModel(ctx, model)
	ctxhelper.PutServiceProvider(ctx, serviceProvider)
	ctxhelper.PutPromptTemplate(ctx, prompt)
	ctxhelper.PutSession(ctx, session)

	// model name will be set by specific context-xxx filters

	// save to db
	return f.saveContextToAudit(pr)
}

func (f *Context) saveContextToAudit(pr *httputil.ProxyRequest) error {
	ctx := pr.Out.Context()
	if model, ok := ctxhelper.GetModel(ctx); ok {
		audithelper.Note(ctx, "model_id", model.Id)
	}
	if session, _ := ctxhelper.GetSession(ctx); session != nil {
		audithelper.Note(ctx, "session_id", session.Id)
	}
	audithelper.Note(ctx, "source", vars.GetFromHeader(pr.Out.Header, vars.XAIProxySource))

	return nil
}
