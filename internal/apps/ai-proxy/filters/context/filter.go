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
	"sync"

	"github.com/erda-project/erda-infra/base/logs"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
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
	ak := vars.TrimBearer(infor.Header().Get("Authorization"))
	if ak == "" {
		http.Error(w, "Authorization is required", http.StatusUnauthorized)
		return reverseproxy.Intercept, nil
	}
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

	// find model
	headerModelId := infor.Header().Get(vars.XAIProxyModelId)
	if headerModelId == "" {
		http.Error(w, fmt.Sprintf("header %s is required", vars.XAIProxyModelId), http.StatusBadRequest)
		return reverseproxy.Intercept, nil
	}
	model, err := q.ModelClient().Get(ctx, &modelpb.ModelGetRequest{Id: headerModelId})
	if err != nil {
		l.Errorf("failed to get model, id: %s, err: %v", headerModelId, err)
		http.Error(w, "ModelId is invalid", http.StatusBadRequest)
		return reverseproxy.Intercept, err
	}

	// find provider
	modelProvider, err := q.ModelProviderClient().Get(ctx, &modelproviderpb.ModelProviderGetRequest{Id: model.ProviderId})
	if err != nil {
		l.Errorf("failed to get model provider, id: %s, err: %v", model.ProviderId, err)
		http.Error(w, "ModelProviderId is invalid", http.StatusBadRequest)
		return reverseproxy.Intercept, err
	}

	// store data to context
	m.Store(vars.MapKeyClient{}, &clientPagingResult.List[0])
	m.Store(vars.MapKeyModel{}, &model)
	m.Store(vars.MapKeyModelProvider{}, &modelProvider)

	return reverseproxy.Continue, nil
}
