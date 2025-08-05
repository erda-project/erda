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
	"strings"
	"sync"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
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
		l = ctxhelper.MustGetLogger(pr.In.Context())
		q = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
		m = ctx.Value(ctxhelper.CtxKeyMap{}).(*sync.Map)
	)

	// find client
	var client *clientpb.Client
	ak := vars.TrimBearer(pr.Out.Header.Get(httperrorutil.HeaderKeyAuthorization))
	if ak == "" {
		return http_error.NewHTTPError(http.StatusUnauthorized, "Authorization is required")
	}
	if strings.HasPrefix(ak, client_token.TokenPrefix) {
		tokenPagingResp, err := q.ClientTokenClient().Paging(ctx, &clienttokenpb.ClientTokenPagingRequest{
			PageSize: 1,
			PageNum:  1,
			Token:    ak,
		})
		if err != nil || tokenPagingResp.Total < 1 {
			l.Errorf("failed to get client token, token: %s, err: %v", ak, err)
			return http_error.NewHTTPError(http.StatusForbidden, "failed to get client token")
		}
		token := tokenPagingResp.List[0]
		clientResp, err := q.ClientClient().Get(ctx, &clientpb.ClientGetRequest{ClientId: token.ClientId})
		if err != nil {
			l.Errorf("failed to get client, id: %s, err: %v", tokenPagingResp.List[0].ClientId, err)
			return http_error.NewHTTPError(http.StatusForbidden, "Authorization is invalid")
		}
		client = clientResp
		m.Store(vars.MapKeyClientToken{}, token)
	} else {
		clientPagingResult, err := q.ClientClient().Paging(ctx, &clientpb.ClientPagingRequest{
			AccessKeyIds: []string{ak},
			PageNum:      1,
			PageSize:     1,
		})
		if err != nil || clientPagingResult.Total < 1 {
			l.Errorf("failed to get client, access_key_id: %s, err: %v", ak, err)
			return http_error.NewHTTPError(http.StatusForbidden, "Authorization is invalid")
		}
		client = clientPagingResult.List[0]
	}

	// session
	var session *sessionpb.Session
	headerSessionId := pr.Out.Header.Get(vars.XAIProxySessionId)
	if headerSessionId != "" && headerSessionId != vars.UIValueUndefined {
		_session, err := q.SessionClient().Get(ctx, &sessionpb.SessionGetRequest{Id: headerSessionId})
		if err != nil {
			l.Errorf("failed to get session, id: %s, err: %v", headerSessionId, err)
			return http_error.NewHTTPError(http.StatusBadRequest, "SessionId is invalid")
		}
		session = _session
	}

	// find model
	model, err := findModel(pr.In, pr.In.Context(), client)
	if err != nil {
		l.Errorf("failed to request model, err: %v", err)
		return http_error.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Model is invalid: %v", err))
	}

	// find provider
	var modelProvider *modelproviderpb.ModelProvider
	modelProvider, err = q.ModelProviderClient().Get(ctx, &modelproviderpb.ModelProviderGetRequest{Id: model.ProviderId})
	if err != nil {
		l.Errorf("failed to get model provider, id: %s, err: %v", model.ProviderId, err)
		return http_error.NewHTTPError(http.StatusBadRequest, "ModelProviderId is invalid")
	}

	// find prompt
	headerPromptId := pr.Out.Header.Get(vars.XAIProxyPromptId)
	var prompt *promptpb.Prompt
	if headerPromptId != "" {
		_prompt, err := q.PromptClient().Get(ctx, &promptpb.PromptGetRequest{Id: headerPromptId})
		if err != nil {
			l.Errorf("failed to get prompt, id: %s, err: %v", headerPromptId, err)
			return http_error.NewHTTPError(http.StatusBadRequest, "Prompt is invalid")
		}
		prompt = _prompt
	}

	// store data to context
	m.Store(vars.MapKeyClient{}, client)
	m.Store(vars.MapKeyModel{}, model)
	m.Store(vars.MapKeyModelProvider{}, modelProvider)
	m.Store(vars.MapKeyPromptTemplate{}, prompt)
	m.Store(vars.MapKeySession{}, session)

	// model name will be set by specific context-xxx filters

	// save to db
	return f.saveContextToAudit(pr)
}

func getModelTypeByRequest(routerPath string) (modelpb.ModelType, bool) {
	if strutil.HasPrefixes(routerPath, common.RequestPathPrefixV1ChatCompletions, common.RequestPathPrefixV1Completions) {
		return modelpb.ModelType_text_generation, true
	}
	if strutil.HasPrefixes(routerPath, common.RequestPathPrefixV1Images) {
		return modelpb.ModelType_image, true
	}
	if strutil.HasPrefixes(routerPath, common.RequestPathPrefixV1Audio) {
		return modelpb.ModelType_audio, true
	}
	if strutil.HasPrefixes(routerPath, common.RequestPathPrefixV1Embeddings) {
		return modelpb.ModelType_embedding, true
	}
	if strutil.HasPrefixes(routerPath, common.RequestPathPrefixV1Moderations) {
		return modelpb.ModelType_text_moderation, true
	}
	if strutil.HasPrefixes(routerPath, common.RequestPathPrefixV1Assistants, common.RequestPathPrefixV1Threads) {
		return modelpb.ModelType_assistant, true
	}
	return -1, false
}

func (f *Context) saveContextToAudit(pr *httputil.ProxyRequest) error {
	auditRecID, ok := ctxhelper.GetAuditID(pr.Out.Context())
	if !ok || auditRecID == "" {
		return nil
	}

	var updateReq pb.AuditUpdateRequestAfterBasicContextParsed
	updateReq.AuditId = auditRecID

	// client id
	client, _ := ctxhelper.GetClient(pr.Out.Context())
	updateReq.ClientId = client.Id
	// model id
	model, _ := ctxhelper.GetModel(pr.Out.Context())
	updateReq.ModelId = model.Id
	// session id
	session, _ := ctxhelper.GetSession(pr.Out.Context())
	if session != nil {
		updateReq.SessionId = session.Id
	}

	// biz source
	updateReq.BizSource = vars.GetFromHeader(pr.Out.Header, vars.XAIProxySource)
	// operation id
	updateReq.OperationId = pr.Out.Method + " " + pr.Out.URL.Path

	// set from client token
	setUserInfoFromClientToken(pr, &updateReq)

	// update audit into db
	_, err := ctxhelper.MustGetDBClient(pr.Out.Context()).AuditClient().UpdateAfterBasicContextParsed(pr.Out.Context(), &updateReq)
	if err != nil {
		// log it
		l := ctxhelper.MustGetLogger(pr.Out.Context())
		l.Errorf("failed to update audit: %v", err)
	}
	return nil
}

func setUserInfoFromClientToken(pr *httputil.ProxyRequest, updateReq *pb.AuditUpdateRequestAfterBasicContextParsed) {
	clientToken, ok := ctxhelper.GetClientToken(pr.Out.Context())
	if !ok || clientToken == nil {
		return
	}
	meta := metadata.FromProtobuf(clientToken.Metadata)
	metaCfg := metadata.Config{IgnoreCase: true}
	updateReq.DingtalkStaffId = meta.MustGetValueByKey(vars.XAIProxyDingTalkStaffID, metaCfg)
	updateReq.Email = meta.MustGetValueByKey(vars.XAIProxyEmail, metaCfg)
	updateReq.IdentityJobNumber = meta.MustGetValueByKey(vars.XAIProxyJobNumber, metaCfg)
	updateReq.Username = meta.MustGetValueByKey(vars.XAIProxyName, metaCfg)
	updateReq.IdentityPhoneNumber = meta.MustGetValueByKey(vars.XAIProxyPhone, metaCfg)
	if vars.GetFromHeader(pr.Out.Header, vars.XAIProxySource) == "" { // use token's client's name
		client, ok := ctxhelper.GetClient(pr.Out.Context())
		if ok {
			updateReq.BizSource = client.Name
		}
	}
}
