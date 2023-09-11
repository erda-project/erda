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
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
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
	ak := vars.TrimBearer(infor.Header().Get("Authorization"))
	if ak == "" {
		http.Error(w, "Authorization is required", http.StatusUnauthorized)
		return reverseproxy.Intercept, nil
	}
	// try to remove Bearer
	ak = strings.TrimPrefix(ak, "Bearer ")
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

	// find model
	var model *modelpb.Model
	var session *sessionpb.Session
	// get from session if exists
	headerSessionId := infor.Header().Get(vars.XAIProxySessionId)
	headerModelId := infor.Header().Get(vars.XAIProxyModelId)
	if headerSessionId != "" {
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
		modelType, ok := getModelTypeByRequest(infor)
		if !ok {
			l.Errorf("failed to judge model type by request path")
			http.Error(w, "ModelType is invalid", http.StatusBadRequest)
			return reverseproxy.Intercept, err
		}
		defaultModelId, ok := clientMeta.Public.GetDefaultModelIdByModelType(modelType)
		if !ok {
			l.Errorf("failed to get client's default model")
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
	modelProvider, err := q.ModelProviderClient().Get(ctx, &modelproviderpb.ModelProviderGetRequest{Id: model.ProviderId})
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
	m.Store(vars.MapKeyPrompt{}, prompt)
	m.Store(vars.MapKeySession{}, session)

	return reverseproxy.Continue, nil
}

func getModelTypeByRequest(infor reverseproxy.HttpInfor) (modelpb.ModelType, bool) {
	if strutil.HasPrefixes(infor.URL().Path, "/v1/chat/completions", "/v1/completions") {
		return modelpb.ModelType_text_generation, true
	}
	if strutil.HasPrefixes(infor.URL().Path, "/v1/images") {
		return modelpb.ModelType_image, true
	}
	if strutil.HasPrefixes(infor.URL().Path, "/v1/audio") {
		return modelpb.ModelType_audio, true
	}
	if strutil.HasPrefixes(infor.URL().Path, "/v1/embeddings") {
		return modelpb.ModelType_embedding, true
	}
	if strutil.HasPrefixes(infor.URL().Path, "/v1/moderations") {
		return modelpb.ModelType_text_moderation, true
	}
	return -1, false
}
