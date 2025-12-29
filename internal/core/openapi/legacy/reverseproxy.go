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

package legacy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api"
	apispec "github.com/erda-project/erda/internal/core/openapi/legacy/api/spec"
	"github.com/erda-project/erda/internal/core/openapi/legacy/auth"
	"github.com/erda-project/erda/internal/core/openapi/legacy/monitor"
	"github.com/erda-project/erda/internal/core/openapi/legacy/proxy"
	phttp "github.com/erda-project/erda/internal/core/openapi/legacy/proxy/http"
	"github.com/erda-project/erda/internal/core/openapi/legacy/proxy/ws"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
)

type ReverseProxyWithAuth struct {
	httpProxy http.Handler
	wsProxy   http.Handler
	auth      *auth.Auth
	bundle    *bundle.Bundle
	cache     *sync.Map
}

func NewReverseProxyWithAuth(auth *auth.Auth, bundle *bundle.Bundle) (http.Handler, error) {
	director := proxy.NewDirector()
	httpProxy := phttp.NewReverseProxyWithCustom(director)
	wsProxy := ws.NewReverseProxyWithCustom(director)
	return &ReverseProxyWithAuth{httpProxy: httpProxy, wsProxy: wsProxy, auth: auth, bundle: bundle, cache: &sync.Map{}}, nil
}

func (r *ReverseProxyWithAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	spec := api.API.Find(req)
	if spec == nil {
		errStr := fmt.Sprintf("not found path: %v", req.URL)
		logrus.Error(errStr)
		http.Error(rw, errStr, 404)
		return
	}

	if authr := r.auth.Auth(spec, req); authr.Code != auth.AuthSucc {
		errStr := fmt.Sprintf("auth failed: %v", authr.Detail)
		logrus.Error(errStr)
		http.Error(rw, errStr, authr.Code)
		return
	}

	if refresh := auth.GetSessionRefresh(req.Context()); refresh != nil {
		if writer, ok := r.auth.CredStore.(domain.RefreshWriter); ok {
			if err := writer.WriteRefresh(rw, req, refresh); err != nil {
				logrus.Warnf("failed to refresh refresh %v", err)
			}
		}
	}

	switch spec.Scheme {
	case apispec.HTTP:
		monitor.Notify(monitor.Info{
			Tp:     monitor.APIInvokeCount,
			Detail: spec.Path.String(),
		})
		start := time.Now()
		if !spec.ChunkAPI && spec.Audit != nil {
			reqBody, err := io.ReadAll(req.Body)
			errStr := fmt.Sprintf("read body failed: %v", err)
			if err != nil {
				logrus.Error(errStr)
				http.Error(rw, errStr, http.StatusBadRequest)
				return
			}
			c := context.WithValue(req.Context(), "reqBody", io.NopCloser(bytes.NewReader(reqBody)))
			c = context.WithValue(c, "bundle", r.bundle)
			c = context.WithValue(c, "beginTime", time.Now())
			c = context.WithValue(c, "cache", r.cache)
			req = req.WithContext(c)
			req.Body = io.NopCloser(bytes.NewReader(reqBody))
			r.httpProxy.ServeHTTP(rw, req)
		} else {
			r.httpProxy.ServeHTTP(rw, req)
		}

		elapsed := time.Since(start)
		monitor.Notify(monitor.Info{
			Tp:     monitor.APIInvokeDuration,
			Detail: spec.Path.String(),
			Value:  elapsed.Nanoseconds() / 1000000, // ms
		})
	case apispec.WS:
		r.wsProxy.ServeHTTP(rw, req)
	default:
		errStr := fmt.Sprintf("not support scheme: %v", spec.Scheme)
		logrus.Error(errStr)
		http.Error(rw, errStr, http.StatusBadRequest)
		return
	}
}
