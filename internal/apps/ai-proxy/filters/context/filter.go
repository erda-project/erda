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
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
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
		providers   = ctx.Value(vars.CtxKeyProviders{}).(provider.Providers)
		credentials []*models.AIProxyCredentials
		appKey      = strings.TrimPrefix(infor.Header().Get("Authorization"), "Bearer ")
	)
	if err = q.Find(&credentials, map[string]any{"access_key_id": appKey}).Error; err != nil || len(credentials) == 0 {
		l.Errorf("failed to Find credentials, access_key_id: %s, err: %v", appKey, err)
		http.Error(w, "Authorization is invalid", http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}

	// find valid credential
	var (
		credential *models.AIProxyCredentials
		match      = func(*models.AIProxyCredentials) bool { return true }
	)
	// if a provider is specified in the request header, refactor the function match
	if providerName := infor.Header().Get(vars.XAIProxyProvider); providerName != "" {
		providerInstanceId := infor.Header().Get(vars.XAIProxyProviderInstance)
		if providerInstanceId == "" {
			providerInstanceId = "default"
		}
		match = func(item *models.AIProxyCredentials) bool {
			return item.Provider == providerName && item.ProviderInstanceId == providerInstanceId
		}
	}
	for _, item := range credentials {
		if item.Enabled && item.ExpiredAt.After(time.Now()) && match(item) {
			credential = item
			break
		}
	}
	if credential == nil {
		http.Error(w, "Authorization is disabled or expired", http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}

	// find provider
	prov, ok := providers.FindProvider(credential.Provider, credential.ProviderInstanceId)
	if !ok {
		http.Error(w, "Provider Not Found", http.StatusNotFound)
		return reverseproxy.Intercept, nil
	}

	// load data to context
	m.Store(vars.MapKeyCredential{}, &credential)
	m.Store(vars.MapKeyProvider{}, prov)

	return reverseproxy.Continue, nil
}
