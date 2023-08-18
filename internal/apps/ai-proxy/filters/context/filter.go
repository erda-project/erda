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
	"net/http"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
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
		l           = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
		q           = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO).Q()
		m           = ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map)
		prov        models.AIProxyProviders
		credentials models.AIProxyCredentialsList
		appKey      = vars.TrimBearer(infor.Header().Get("Authorization"))
	)
	total, err := (&credentials).Pager(q).
		Where(credentials.FieldAccessKeyID().Equal(appKey)).
		Paging(-1, -1)
	if err != nil {
		l.Errorf("failed to list credentials, access_key_id: %s, err: %v", appKey, err)
		return reverseproxy.Intercept, err
	}
	if total == 0 {
		l.Debugf("no credential with the access_key_id: %s", appKey)
		http.Error(w, "Authorization is invalid", http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}

	// find valid credential
	var (
		credential *models.AIProxyCredentials
		match      = func(*models.AIProxyCredentials) bool { return true }
	)
	// if a provider is specified in the request header, refactor the function match
	if providerName := infor.Header().Get(vars.XAIProxyProviderName); providerName != "" {
		providerInstanceId := infor.Header().Get(vars.XAIProxyProviderInstanceId)
		if providerInstanceId == "" {
			providerInstanceId = "default"
		}
		match = func(item *models.AIProxyCredentials) bool {
			return item.ProviderName == providerName && item.ProviderInstanceID == providerInstanceId
		}
	}
	// 从 credentials 中取出匹配的 credential: 启用状态为 true && 未过期 && 能匹配上 provider
	for _, item := range credentials {
		if item.Enabled && item.ExpiredAt.After(time.Now()) && match(item) {
			credential = item
			break
		}
	}
	if credential == nil {
		http.Error(w, "Authorization is disabled or expired or not matches the specified provider", http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}

	// 取出这个 credential 对应的 provider
	if ok, _ := (&prov).Retriever(q).Where(
		prov.FieldName().Equal(credential.ProviderName),
		prov.FieldInstanceID().Equal(credential.ProviderInstanceID),
	).Get(); !ok {
		http.Error(w, "ProviderName Not Found", http.StatusBadRequest)
		return reverseproxy.Intercept, nil
	}

	// store data to context
	m.Store(vars.MapKeyCredential{}, &credential)
	m.Store(vars.MapKeyProvider{}, &prov)

	return reverseproxy.Continue, nil
}
