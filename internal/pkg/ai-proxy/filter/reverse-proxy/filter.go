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

package reverse_proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name       = "reverse-proxy"
	ToProvider = "provider"
	ToURL      = "url"
)

var (
	_    filter.Filter = (*ReverseProxy)(nil)
	Type               = reflect.TypeOf((*ReverseProxy)(nil))
)

func init() {
	filter.Register(Name, New)
}

type ReverseProxy struct {
	Config *Config

	cachedURL    url.URL
	cachedHeader http.Header
}

func New(config json.RawMessage) (filter.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	if cfg.InstanceId == "" {
		cfg.InstanceId = "default"
	}
	return &ReverseProxy{Config: &cfg}, nil
}

func (f *ReverseProxy) OnHttpRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	switch f.Config.To {
	case "", ToProvider:
		return f.onHttpRequestToProvider(ctx, w, r)
	default:
		return f.onHttpRequestToURL(ctx, w, r)
	}
}

func (f *ReverseProxy) OnHttpResponse(_ context.Context, _ *http.Response, _ *http.Request) (filter.Signal, error) {
	return filter.Continue, nil
}

func (f *ReverseProxy) onHttpRequestToProvider(ctx context.Context, w http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	// retrieve contexts: logger, route, providers, provider, operation, filters
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger)
	rout, ok := ctx.Value(filter.RouteCtxKey{}).(*route.Route)
	if !ok {
		w.Header().Set("server", "ai-proxy/erda")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "failed to retrieve route info",
		})
		return filter.Intercept, nil
	}

	providers, ok := ctx.Value(filter.ProvidersCtxKey{}).(provider.Providers)
	if !ok {
		w.Header().Set("server", "ai-proxy/erda")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "failed to retrieve providers info",
		})
		return filter.Intercept, nil
	}
	prov, ok := providers.FindProvider(f.Config.Provider, f.Config.InstanceId)
	if !ok {
		w.Header().Set("server", "ai-proxy/erda")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    "no such provider",
			"provider": f.Config.Provider,
		})
		return filter.Intercept, nil
	}
	filter.WithValue(ctx, filter.ProviderCtxKey{}, prov)

	operation, ok := prov.FindAPI(f.Config.OperationId, f.Config.Path, f.mappedMethod(rout.Method))
	if !ok {
		w.Header().Set("server", "ai-proxy/erda")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "API not found",
			"provider":    prov.Name,
			"operationId": f.Config.OperationId,
			"apiPath":     f.Config.Path,
		})
		return filter.Intercept, nil
	}
	filter.WithValue(ctx, filter.OperationCtxKey{}, operation)
	filters, ok := ctx.Value(filter.FiltersCtxKey{}).([]filter.Filter)
	if !ok {
		w.Header().Set("server", "ai-proxy/erda")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "failed to retrieve filters info",
		})
	}

	var totalFilters int
	var director = func(r *http.Request) {
		defer func() {
			f.cachedURL = *r.URL
			f.cachedHeader = r.Header
			var m = map[string]any{
				"scheme":     r.URL.Scheme,
				"host":       r.Host,
				"uri":        r.URL.RequestURI(),
				"headers":    r.Header,
				"remoteAddr": r.RemoteAddr,
			}
			if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				data, err := io.ReadAll(r.Body)
				if err != nil {
					l.Errorf(`[ReverseProxy] failed to io.ReadAll(r.Body), err: %v`, err)
				} else {
					defer func() {
						r.Body = io.NopCloser(bytes.NewReader(data))
					}()
					var buf bytes.Buffer
					_ = json.Indent(&buf, data, "  ", "  ")
					m["body"] = buf.String()
				}
			}
			l.Infof("[ReverseProxy] request info:\n%s\n", strutil.TryGetYamlStr(m))
		}()

		r.URL.Scheme = "http"
		if prov.Scheme != "" {
			r.URL.Scheme = prov.Scheme
		}
		if f.Config.Scheme != "" {
			r.URL.Scheme = f.Config.Scheme
		}
		r.URL.Host = prov.GetHost()
		r.Host = prov.GetHost()
		r.Header.Set("host", prov.GetHost())
		if len(f.Config.MethodsMap) > 0 {
			if method, ok := f.Config.MethodsMap[r.Method]; ok && method != "" {
				r.Method = method
			}
		}
		//_ = api // todo: rewrite path
		// Rewrite provider credentials and add credential information according to the configuration
		// if the original request does not carry credential information
		// todo: 对每一个 provider 重写凭证的方式都抽成函数或进行抽象
		switch prov.Name {
		case provider.ChatGPTv1:
			if appKey := prov.GetAppKey(); appKey != "" && r.Header.Get("Authorization") == "" {
				r.Header.Set("Authorization", "Bearer "+appKey)
			}
			if org := prov.GetOrganization(); org != "" && r.Header.Get("OpenAI-Organization") == "" {
				r.Header.Set("OpenAI-Organization", org)
			}
		default:
			// pass
		}
	}
	var modifyResponse = func(p *http.Response) error {
		if p.StatusCode < http.StatusOK || p.StatusCode > http.StatusMultipleChoices {
			l.Errorf(`[ReverseProxy] failed to request upstream server,
	scheme: %s
	host: 	%s
	uri: 	%s
	status:	%s
	`, f.cachedURL.Scheme, f.cachedURL.Host, f.cachedURL.RequestURI(), p.Status)
		}
		for i := totalFilters - 1; i >= 0; i-- {
			if reflect.TypeOf(filters[i]) == Type {
				continue
			}
			if on, ok := filters[i].(filter.ResponseFilter); ok {
				l.Debugf(`[ReverseProxy] %T.OnHttpResponse called`, on)
				signal, err := on.OnHttpResponse(ctx, p)
				if err != nil {
					l.Errorf("[ReverseProxy][ModifyResponse] failed to OnHttpResponse, signal: %v, err: %v", signal, err)
					continue
				}
				if signal != filter.Continue {
					continue
				}
			}
		}
		return nil
	}

	for i := 0; i < len(filters); i++ {
		totalFilters++
		if reflect.TypeOf(filters[i]) == Type {
			continue
		}
		if on, ok := filters[i].(filter.RequestHeaderFilter); ok {
			l.Debugf(`[ReverseProxy] %T.OnHttpRequestHeader called`, on)
			switch signal, err := on.OnHttpRequestHeader(ctx, r.Header); {
			case err == nil && signal == filter.Continue:
			case err != nil:
				ResponseErrorHandler(w, r, err)
				fallthrough
			default:
				director = doNotDirect
				break
			}
		}
		if on, ok := filters[i].(filter.RequestBodyCopyFilter); ok {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				l.Errorf(`[ReverseProxy] failed to io.ReadAll(r.Body)`)
				ResponseErrorHandler(w, r, err)
				director = doNotDirect
				break
			}
			var buf = bytes.NewReader(data)
			r.Body = io.NopCloser(bytes.NewReader(data))
			l.Debugf(`[ReverseProxy] %T.OnHttpRequestBodyCopy called`, on)
			switch signal, err := on.OnHttpRequestBodyCopy(ctx, buf); {
			case err == nil && signal == filter.Continue:
			case err != nil:
				ResponseErrorHandler(w, r, err)
				fallthrough
			default:
				director = doNotDirect
				break
			}
		}
		if on, ok := filters[i].(filter.RequestFilter); ok {
			l.Debugf(`[ReverseProxy] %T.OnHttpRequest called`, on)
			switch signal, err := on.OnHttpRequest(ctx, w, r); ok {
			case err == nil && signal == filter.Continue:
			case err != nil:
				ResponseErrorHandler(w, r, err)
				fallthrough
			default:
				director = doNotDirect
				break
			}
		}
	}

	(&httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
		ErrorHandler:   ResponseErrorHandler,
	}).ServeHTTP(w, r)

	return filter.Continue, nil
}

func (f *ReverseProxy) onHttpRequestToURL(ctx context.Context, w http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	w.Header().Set("server", "ai-proxy/erda")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "not implement to reverse proxy to url",
	})
	return filter.Intercept, nil
}

func (f *ReverseProxy) mappedMethod(routeMethod string) string {
	if len(f.Config.MethodsMap) == 0 {
		return routeMethod
	}
	method, ok := f.Config.MethodsMap[routeMethod]
	if !ok || !map[string]bool{
		http.MethodGet:     true,
		http.MethodHead:    true,
		http.MethodPost:    true,
		http.MethodPut:     true,
		http.MethodPatch:   true,
		http.MethodDelete:  true,
		http.MethodConnect: true,
		http.MethodOptions: true,
		http.MethodTrace:   true,
	}[strings.ToUpper(method)] {
		return routeMethod
	}
	return strings.ToUpper(method)
}

type Config struct {
	To          string            `json:"to" yaml:"to"`
	Provider    string            `json:"provider" yaml:"provider"`
	InstanceId  string            `json:"instanceId" yaml:"instanceId"`
	OperationId string            `json:"operationId" yaml:"operationId"`
	Scheme      string            `json:"scheme" yaml:"scheme"`
	Host        string            `json:"host" yaml:"host"`
	Path        string            `json:"path" yaml:"path"`
	MethodsMap  map[string]string `json:"methodsMap" yaml:"methodsMap"`
}

func ResponseErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("server", "ai-proxy/erda")
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func doNotDirect(r *http.Request) {}