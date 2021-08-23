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

package collector

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/secret"
	"github.com/erda-project/erda/pkg/secret/validator"
)

var bdl = bundle.New(bundle.WithCoreServices())

type signAuthConfig struct {
	SyncInterval    time.Duration `file:"sync_interval" default:"3m" desc:"sync access key info from remote"`
	ExpiredDuration time.Duration `file:"expired_duration" default:"10m" desc:"the max duration of request spent in the network"`
}

type Authenticator struct {
	store  map[string]*model.AccessKey
	mu     sync.RWMutex
	logger logs.Logger
}

func boolPointer(data bool) *bool {
	return &data
}

func (a *Authenticator) syncAccessKey(ctx context.Context) error {
	newStore := make(map[string]*model.AccessKey, len(a.store))

	objs, err := bdl.ListAccessKeys(apistructs.AccessKeyListQueryRequest{
		IsSystem: boolPointer(true),
		Status:   "ACTIVE",
	})
	if err != nil {
		return fmt.Errorf("syncAccessKey.ListAccessKeys failed. err: %w", err)
	}
	for _, obj := range objs {
		newStore[obj.AccessKeyID] = &obj
	}

	a.mu.Lock()
	a.store = newStore
	a.mu.Unlock()
	return nil
}

func (a *Authenticator) getAccessKey(ak string) (*model.AccessKey, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	v, ok := a.store[ak]
	return v, ok
}

func (p *provider) authSignedRequest() httpserver.Interceptor {
	return func(handler func(ctx httpserver.Context) error) func(ctx httpserver.Context) error {
		return func(ctx httpserver.Context) error {
			ak, ok := validator.GetAccessKeyID(ctx.Request())
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "must specify accessKeyID")
			}

			accessKey, ok := p.auth.getAccessKey(ak)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "can't find accessKeyID: "+ak)
			}

			vd := validator.NewHMACValidator(
				secret.AkSkPair{AccessKeyID: ak, SecretKey: accessKey.SecretKey},
				validator.WithMaxExpireInterval(p.Cfg.SignAuth.ExpiredDuration),
			)
			if res := vd.Verify(ctx.Request()); !res.Ok {
				return echo.NewHTTPError(http.StatusUnauthorized, res.Message)
			}
			return handler(ctx)
		}
	}
}
