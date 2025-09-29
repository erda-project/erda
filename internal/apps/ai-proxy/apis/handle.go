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

package apis

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/requestid"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/reverse_proxy"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
	"github.com/erda-project/erda/pkg/common/apis"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

type Interface interface {
	FindBestMatch(method, path string) router_define.RouteMatcher
	GetDBClient() dao.DAO
	GetRichClientHandler() *handler_rich_client.ClientHandler
	SetRichClientHandler(*handler_rich_client.ClientHandler)
	GetCacheManager() cachetypes.Manager
	IsMcpProxy() bool
}

var (
	trySetAuth = func(dao dao.DAO) transport.ServiceOption {
		return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				ctx = ctxhelper.InitCtxMapIfNeed(ctx)
				// check admin key first
				isAdmin, err := akutil.CheckAdmin(ctx, req, dao)
				if err != nil {
					return nil, err
				}
				if isAdmin {
					ctxhelper.PutIsAdmin(ctx, true)
					return h(ctx, req)
				}
				// try set clientId by ak
				clientToken, client, err := akutil.CheckAkOrToken(ctx, req, dao)
				if err != nil {
					return nil, err
				}
				if clientToken != nil {
					ctxhelper.PutClientToken(ctx, clientToken)
				}
				if client != nil {
					ctxhelper.PutClient(ctx, client)
					ctxhelper.PutClientId(ctx, client.Id)
				}
				return h(ctx, req)
			}
		})
	}
	trySetLang = func() transport.ServiceOption {
		return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				ctx = ctxhelper.InitCtxMapIfNeed(ctx)
				lang := apis.GetHeader(ctx, httperrorutil.HeaderKeyAcceptLanguage)
				if len(lang) > 0 {
					ctxhelper.PutAccessLang(ctx, lang)
				}
				return h(ctx, req)
			}
		})
	}
)

func ReverseProxyHandle[T Interface](p T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// inject context
		ctx := ctxhelper.InitCtxMapIfNeed(r.Context())
		r = r.WithContext(ctx)

		// create a request-level Logger
		logger := logrusx.New().Sub("reverse-proxy-api")
		baseLogger := logrusx.New().Sub("reverse-proxy-api")
		reqID := requestid.GetOrGenRequestID(r)
		callID := requestid.GetCallID(r)
		logger.Set("req", reqID).Set("call", callID)
		baseLogger.Set("req", reqID).Set("call", callID)
		baseLogger.Infof("reverse proxy handler: %s %s", r.Method, r.URL.String())
		ctxhelper.PutLogger(ctx, logger)
		ctxhelper.PutLoggerBase(ctx, baseLogger)
		ctxhelper.PutRequestID(ctx, reqID)
		ctxhelper.PutGeneratedCallID(ctx, callID)

		// find best matched route using priority algorithm
		matched := p.FindBestMatch(r.Method, r.URL.Path)
		var matchedRoute *router_define.Route
		if matched != nil {
			matchedRoute = matched.(*router_define.Route)
		}
		if matchedRoute == nil {
			httperror.NewHTTPError(r.Context(), http.StatusNotFound, "no matched route").WriteJSONHTTPError(w)
			return
		}

		// get all route filters
		var requestFilters []filter_define.FilterWithName[filter_define.ProxyRequestRewriter]
		for _, filterConfig := range matchedRoute.RequestFilters {
			creator, ok := filter_define.FilterFactory.RequestFilters[filterConfig.Name]
			if !ok {
				httperror.NewHTTPError(r.Context(), http.StatusInternalServerError, fmt.Sprintf("request filter %s not found", filterConfig.Name)).WriteJSONHTTPError(w)
				return
			}
			fc, err := filterConfig.GetConfigAsRawMessage()
			if err != nil {
				httperror.NewHTTPError(r.Context(), http.StatusInternalServerError, fmt.Sprintf("failed to convert config for filter %s: %v", filterConfig.Name, err)).WriteJSONHTTPError(w)
				return
			}
			f := creator(filterConfig.Name, fc)
			requestFilters = append(requestFilters, filter_define.FilterWithName[filter_define.ProxyRequestRewriter]{Name: filterConfig.Name, Instance: f})
		}
		var responseFilters []filter_define.FilterWithName[filter_define.ProxyResponseModifier]
		for _, filterConfig := range matchedRoute.ResponseFilters {
			creator, ok := filter_define.FilterFactory.ResponseFilters[filterConfig.Name]
			if !ok {
				httperror.NewHTTPError(r.Context(), http.StatusInternalServerError, fmt.Sprintf("response filter %s not found", filterConfig.Name)).WriteJSONHTTPError(w)
				return
			}
			fc, err := filterConfig.GetConfigAsRawMessage()
			if err != nil {
				httperror.NewHTTPError(r.Context(), http.StatusInternalServerError, fmt.Sprintf("failed to convert config for filter %s: %v", filterConfig.Name, err)).WriteJSONHTTPError(w)
				return
			}
			f := creator(filterConfig.Name, fc)
			responseFilters = append(responseFilters, filter_define.FilterWithName[filter_define.ProxyResponseModifier]{Name: filterConfig.Name, Instance: f})
		}

		ctxhelper.PutDBClient(ctx, p.GetDBClient())
		ctxhelper.PutRichClientHandler(ctx, p.GetRichClientHandler())
		ctxhelper.PutPathMatcher(ctx, matchedRoute.GetPathMatcher())
		ctxhelper.PutCacheManager(ctx, p.GetCacheManager())

		proxy := httputil.ReverseProxy{
			Rewrite: reverse_proxy.MyRewrite(w, requestFilters),
			Transport: func() http.RoundTripper {
				if p.IsMcpProxy() {
					return transports.NewMcpTransport()
				}
				return transports.NewRequestFilterGeneratedResponseTransport(&transports.TimerTransport{})
			}(),
			FlushInterval:  -1,
			ErrorLog:       nil,
			BufferPool:     nil,
			ModifyResponse: reverse_proxy.MyResponseModify(w, responseFilters),
			ErrorHandler:   reverse_proxy.MyErrorHandler(),
		}
		proxy.ServeHTTP(w, r)
	}
}
