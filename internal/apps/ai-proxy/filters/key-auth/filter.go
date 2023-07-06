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

package key_auth

import (
	"context"
	"encoding/json"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	Name = "key-auth"
)

var (
	_ reverseproxy.RequestFilter = (*KeyAuth)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type KeyAuth struct {
	Cfg *Config
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, errors.Wrapf(err, "failed to parse config %s for %s", string(config), Name)
	}
	if len(cfg.Invalid) == 0 {
		cfg.Invalid = json.RawMessage("appKey is invalid")
	}
	if len(cfg.Disabled) == 0 {
		cfg.Disabled = json.RawMessage("appKey is disabled")
	}
	if len(cfg.Expired) == 0 {
		cfg.Expired = json.RawMessage("appKey is expired")
	}
	return &KeyAuth{Cfg: &cfg}, nil
}

func (f *KeyAuth) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
	appKey := infor.Header().Get("Authorization")
	appKey = strings.TrimPrefix(appKey, "Bearer ")

	var credential models.AIProxyCredentials
	q := ctx.Value(vars.CtxKeyDAO{}).(dao.DAO).Q()
	if err = q.First(&credential, map[string]any{"access_key_id": appKey}).Error; err != nil {
		l.Errorf("failed to First credential: %v", err)
		http.Error(w, string(f.Cfg.Invalid), http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}
	if !credential.Enabled {
		http.Error(w, string(f.Cfg.Disabled), http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}
	if credential.ExpiredAt.Before(time.Now()) {
		http.Error(w, string(f.Cfg.Expired), http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}
	appKey = "Bearer " + ctx.Value(vars.CtxKeyProvider{}).(*provider.Provider).GetAppKey()
	infor.Header().Set("Authorization", appKey)
	ctx.Value(vars.CtxKeyMap{}).(*sync.Map).Store(vars.CtxKeyCredential{}, &credential)
	return reverseproxy.Continue, nil
}

type Config struct {
	Invalid  json.RawMessage `json:"invalid" yaml:"invalid"`
	Disabled json.RawMessage `json:"disabled" yaml:"disabled"`
	Expired  json.RawMessage `json:"expired" yaml:"expired"`
}
