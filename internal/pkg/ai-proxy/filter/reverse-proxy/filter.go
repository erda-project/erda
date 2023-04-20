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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/route"
)

const (
	Name       = "reverse-proxy"
	ToProvider = "provider"
	ToURL      = "url"
)

var (
	_ filter.Filter = (*ReverseProxy)(nil)
)

func init() {
	filter.Register(Name, New)
}

type ReverseProxy struct {
	Config *Config
}

func New(config json.RawMessage) (filter.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &ReverseProxy{Config: &cfg}, nil
}

func (fil *ReverseProxy) OnHttpRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	switch fil.Config.To {
	case "", ToProvider:
		return fil.onHttpRequestToProvider(ctx, w, r)
	default:
		return fil.onHttpRequestToURL(ctx, w, r)
	}
}

func (fil *ReverseProxy) OnHttpResponse(_ context.Context, _ *http.Response, _ *http.Request) (filter.Signal, error) {
	return filter.Continue, nil
}

func (fil *ReverseProxy) onHttpRequestToProvider(ctx context.Context, w http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	rout, ok := ctx.Value(filter.RouteCtxKey{}).(*route.Route)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "failed to retrieve route info",
		})
		return filter.Intercept, nil
	}
	_ = rout // todo: 似乎这里不需要获取 route 的配置
	providers, ok := ctx.Value(filter.ProvidersCtxKey{}).(provider.Providers)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "failed to retrieve providers info",
		})
		return filter.Intercept, nil
	}
	prov, ok := providers.GetProvider(fil.Config.Provider)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    "no such provider",
			"provider": fil.Config.Provider,
		})
		return filter.Intercept, nil
	}
	api, ok := prov.FindAPI(fil.Config.APIName, fil.Config.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    "API not found",
			"provider": prov.Name,
			"apiName":  fil.Config.APIName,
			"apiPath":  fil.Config.Path,
		})
	}
	filters, ok := ctx.Value(filter.FiltersCtxKey{}).([]filter.Filter)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "failed to retrieve filters info",
		})
	}

	(&httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			if prov.Scheme != "" {
				r.URL.Scheme = prov.Scheme
			}
			if fil.Config.Scheme != "" {
				r.URL.Scheme = fil.Config.Scheme
			}
			r.URL.Host = prov.Host
			r.Host = prov.Host
			r.Header.Set("host", prov.Host)
			if len(fil.Config.MethodsMap) > 0 {
				if method, ok := fil.Config.MethodsMap[r.Method]; ok && method != "" {
					r.Method = method
				}
			}
			_ = api // todo: rewrite path
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
			prov.GetAppKey()
		},
		ModifyResponse: func(response *http.Response) error {
			for i := len(filters) - 1; i >= 0; i-- {
				signal, err := filters[i].OnHttpResponse(ctx, response, r)
				if err != nil {
					logrus.WithError(err).WithField("signal", signal).Errorln("failed to OnHttpResponse")
					return err
				}
				if signal != filter.Continue {
					return nil
				}
			}
			return nil
		},
		ErrorHandler: ResponseErrorHandler,
	}).ServeHTTP(w, r)

	return filter.Continue, nil
}

func (fil *ReverseProxy) onHttpRequestToURL(ctx context.Context, w http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "not implement to reverse proxy to url",
	})
	return filter.Intercept, nil
}

type Config struct {
	To         string            `json:"to" yaml:"to"`
	Provider   string            `json:"provider" yaml:"provider"`
	APIName    string            `json:"APIName" yaml:"apiName"`
	Scheme     string            `json:"scheme" yaml:"scheme"`
	Host       string            `json:"host" yaml:"host"`
	Path       string            `json:"path" yaml:"path"`
	MethodsMap map[string]string `json:"methodsMap" yaml:"methodsMap"`
}

func ResponseErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("server", "ai-proxy/erda")
	w.WriteHeader(http.StatusBadGateway)
	logrus.Infof("ResponseErrorHandler: %v", err)
	data, _ := json.Marshal(map[string]any{
		"error": err,
	})
	logrus.Infof("ResponseErrorHandler marshaled err: %s", string(data))
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err,
	})
}
