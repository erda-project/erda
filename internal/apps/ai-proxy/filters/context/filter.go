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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/erda-project/erda-infra/base/logs"
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
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "context"
)

var (
	_ reverseproxy.RequestFilter = (*Context)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Context struct {
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &Context{}, nil
}

func (f *Context) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var (
		l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
		q = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
		m = ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map)
	)
	// find client
	var client *clientpb.Client
	ak := vars.TrimBearer(infor.Header().Get(httputil.HeaderKeyAuthorization))
	if ak == "" {
		http.Error(w, "Authorization is required", http.StatusUnauthorized)
		return reverseproxy.Intercept, nil
	}
	if strings.HasPrefix(ak, client_token.TokenPrefix) {
		tokenPagingResp, err := q.ClientTokenClient().Paging(ctx, &clienttokenpb.ClientTokenPagingRequest{
			PageSize: 1,
			PageNum:  1,
			Token:    ak,
		})
		if err != nil || tokenPagingResp.Total < 1 {
			l.Errorf("failed to get client token, token: %s, err: %v", ak, err)
			http.Error(w, "Authorization is invalid", http.StatusForbidden)
			return reverseproxy.Intercept, err
		}
		token := tokenPagingResp.List[0]
		clientResp, err := q.ClientClient().Get(ctx, &clientpb.ClientGetRequest{ClientId: token.ClientId})
		if err != nil {
			l.Errorf("failed to get client, id: %s, err: %v", tokenPagingResp.List[0].ClientId, err)
			http.Error(w, "Authorization is invalid", http.StatusForbidden)
			return reverseproxy.Intercept, err
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
			http.Error(w, "Authorization is invalid", http.StatusForbidden)
			return reverseproxy.Intercept, err
		}
		client = clientPagingResult.List[0]
	}

	// find model
	var model *modelpb.Model
	var modelProvider *modelproviderpb.ModelProvider
	var session *sessionpb.Session
	// get from session if exists
	headerSessionId := infor.Header().Get(vars.XAIProxySessionId)
	headerModelId := infor.Header().Get(vars.XAIProxyModelId)
	if headerSessionId != "" && headerSessionId != vars.UIValueUndefined {
		_session, err := q.SessionClient().Get(ctx, &sessionpb.SessionGetRequest{Id: headerSessionId})
		if err != nil {
			l.Errorf("failed to get session, id: %s, err: %v", headerSessionId, err)
			http.Error(w, "SessionId is invalid", http.StatusBadRequest)
			return reverseproxy.Intercept, err
		}
		session = _session
		if session.ModelId != "" {
			sessionModel, err := q.ModelClient().Get(ctx, &modelpb.ModelGetRequest{Id: session.ModelId})
			if err != nil {
				l.Errorf("failed to get model, id: %s, err: %v", session.ModelId, err)
				http.Error(w, "ModelId is invalid", http.StatusBadRequest)
				return reverseproxy.Intercept, err
			}
			model = sessionModel
		}
	} else if headerModelId != "" {
		// get from model header
		if headerModelId == "" {
			http.Error(w, fmt.Sprintf("header %s is required", vars.XAIProxyModelId), http.StatusBadRequest)
			return reverseproxy.Intercept, nil
		}
		headerModel, err := q.ModelClient().Get(ctx, &modelpb.ModelGetRequest{Id: headerModelId})
		if err != nil {
			l.Errorf("failed to get model, id: %s, err: %v", headerModelId, err)
			http.Error(w, "ModelId is invalid", http.StatusBadRequest)
			return reverseproxy.Intercept, err
		}
		model = headerModel
	}
	if model == nil {
		// get client default model
		clientPbMeta := metadata.FromProtobuf(client.Metadata)
		clientMeta, err := clientPbMeta.ToClientMeta()
		if err != nil {
			l.Errorf("failed to get client meta, err: %v", err)
			http.Error(w, "Client meta is invalid", http.StatusBadRequest)
			return reverseproxy.Intercept, err
		}
		// judge by model type
		modelType, ok := getModelTypeByRequest(common.GetRequestRoutePath(ctx))
		if !ok {
			l.Errorf("failed to judge model type by request path: %s", infor.URL().Path)
			http.Error(w, fmt.Sprintf("please add http header `X-AI-Proxy-Model-ID` explicitly for path: %s", infor.URL().Path), http.StatusBadRequest)
			return reverseproxy.Intercept, err
		}
		defaultModelId, ok := clientMeta.Public.GetDefaultModelIdByModelType(modelType)
		if !ok {
			l.Errorf("failed to get client's default model, modelType: %s", modelType.String())
			http.Error(w, "Client's default model is invalid", http.StatusBadRequest)
			return reverseproxy.Intercept, err
		}
		defaultModel, err := q.ModelClient().Get(ctx, &modelpb.ModelGetRequest{Id: defaultModelId})
		if err != nil {
			l.Errorf("failed to get model, id: %s, err: %v", defaultModelId, err)
			http.Error(w, "ModelId is invalid", http.StatusBadRequest)
			return reverseproxy.Intercept, err
		}
		model = defaultModel
	}

	// find provider
	modelProvider, err = q.ModelProviderClient().Get(ctx, &modelproviderpb.ModelProviderGetRequest{Id: model.ProviderId})
	if err != nil {
		l.Errorf("failed to get model provider, id: %s, err: %v", model.ProviderId, err)
		http.Error(w, "ModelProviderId is invalid", http.StatusBadRequest)
		return reverseproxy.Intercept, err
	}

	// find prompt
	headerPromptId := infor.Header().Get(vars.XAIProxyPromptId)
	var prompt *promptpb.Prompt
	if headerPromptId != "" {
		_prompt, err := q.PromptClient().Get(ctx, &promptpb.PromptGetRequest{Id: headerPromptId})
		if err != nil {
			l.Errorf("failed to get prompt, id: %s, err: %v", headerPromptId, err)
			http.Error(w, "PromptId is invalid", http.StatusBadRequest)
			return reverseproxy.Intercept, err
		}
		prompt = _prompt
	}

	// store data to context
	m.Store(vars.MapKeyClient{}, client)
	m.Store(vars.MapKeyModel{}, model)
	m.Store(vars.MapKeyModelProvider{}, modelProvider)
	m.Store(vars.MapKeyPromptTemplate{}, prompt)
	m.Store(vars.MapKeySession{}, session)

	// save to db
	return f.saveContextToAudit(ctx, w, infor)
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

func (f *Context) saveContextToAudit(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	auditRecID, ok := ctxhelper.GetAuditID(ctx)
	if !ok || auditRecID == "" {
		return
	}

	var updateReq pb.AuditUpdateRequestAfterBasicContextParsed
	updateReq.AuditId = auditRecID

	// client id
	client, _ := ctxhelper.GetClient(ctx)
	updateReq.ClientId = client.Id
	// model id
	model, _ := ctxhelper.GetModel(ctx)
	updateReq.ModelId = model.Id
	// session id
	session, _ := ctxhelper.GetSession(ctx)
	if session != nil {
		updateReq.SessionId = session.Id
	}

	// biz source
	updateReq.BizSource = vars.GetFromHeader(infor, vars.XAIProxySource)
	// operation id
	updateReq.OperationId = infor.Method() + " " + infor.URL().Path

	// set from client token
	setUserInfoFromClientToken(ctx, infor, &updateReq)

	// update audit into db
	_, err = ctxhelper.MustGetDBClient(ctx).AuditClient().UpdateAfterBasicContextParsed(ctx, &updateReq)
	if err != nil {
		// log it
		l := ctxhelper.GetLogger(ctx)
		l.Errorf("failed to update audit: %v", err)
	}
	return reverseproxy.Continue, nil
}

func setUserInfoFromClientToken(ctx context.Context, infor reverseproxy.HttpInfor, updateReq *pb.AuditUpdateRequestAfterBasicContextParsed) {
	clientToken, ok := ctxhelper.GetClientToken(ctx)
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
	if vars.GetFromHeader(infor, vars.XAIProxySource) == "" { // use token's client's name
		client, ok := ctxhelper.GetClient(ctx)
		if ok {
			updateReq.BizSource = client.Name
		}
	}
}
