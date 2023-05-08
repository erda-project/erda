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

package route

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	ToURL      To = "__url__"
	ToNotFound To = "__not_found__"
)

type Routes []*Route

func (routes Routes) FindRoute(path, method string, header http.Header) (*Route, bool) {
	// todo: 应当改成树形数据结构来存储和查找 route, 不过在 route 数量有限的情形下影响不大
	for _, route := range routes {
		if route.Match(path, method, header) {
			return route, true
		}
	}
	return nil, false
}

type Route struct {
	Path          string          `json:"path" yaml:"path"`
	PathMatcher   string          `json:"pathMatcher" yaml:"pathMatcher"`
	Method        string          `json:"method" yaml:"method"`
	MethodMatcher string          `json:"methodMatcher" yaml:"methodMatcher"`
	HeaderMatcher json.RawMessage `json:"headerMatcher" yaml:"headerMatcher"`
	Router        *Router         `json:"router" yaml:"router"`

	Filters []*reverseproxy.FilterConfig `json:"filters" yaml:"filters"`

	pathMatcher   func(path string) bool
	methodMatcher func(method string) bool
	headerMatcher func(header http.Header) bool

	provider     *provider.Provider
	reverseProxy *reverseproxy.ReverseProxy
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch r.Router.To {
	case ToNotFound:
		http.Error(w, string(ToNotFound), http.StatusNotFound)
		return
	}

	if r.reverseProxy == nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	r.reverseProxy.ServeHTTP(w, req)
}

func (r *Route) PrepareHandler(ctx *reverseproxy.Context, options ...Option) error {
	for _, opt := range options {
		opt(r)
	}
	if err := r.Validate(); err != nil {
		return err
	}

	r.reverseProxy = &reverseproxy.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = r.Router.Scheme
			req.Host = r.Router.Host
			req.URL.Host = r.Router.Host
			req.Header.Set("Host", r.Router.Host)
			req.Header.Set("X-Forwarded-Host", r.Router.Host)
			req.URL.Path = r.rewrite()

			log.Println(strutil.TryGetJsonStr(map[string]any{
				"req.URL.Scheme":                 req.URL.Scheme,
				"req.Host":                       req.Host,
				"req.URL.Host":                   req.URL.Host,
				`req.Header["Host"]`:             req.Header["Host"],
				`req.Header["X-Forwarded-Host"]`: req.Header["X-Forwarded-Host"],
				"req.URL.Path":                   req.URL.Path,
			}))
		},
		FlushInterval:  time.Millisecond * 100,
		ModifyResponse: nil,
		ErrorHandler:   nil,
		Filters:        nil,
		FilterCxt:      ctx,
	}
	for _, filterConfig := range r.Filters {
		filter, err := reverseproxy.MustGetFilterCreator(filterConfig.Name)(filterConfig.Config)
		if err != nil {
			return errors.Wrapf(err, "failed to create filter %s", filterConfig.Name)
		}
		switch filter.(type) {
		case reverseproxy.RequestFilter, reverseproxy.ResponseFilter:
		default:
			return errors.Errorf("filter must be of one type reverseproxy.RequestFilter or reverseproxy.ResponseFilter, got %T", filter)
		}
		r.reverseProxy.Filters = append(r.reverseProxy.Filters, reverseproxy.NamingFilter{Name: filterConfig.Name, Filter: filter})
	}

	return nil
}

func (r *Route) Match(path, method string, header http.Header) bool {
	return r.pathMatcher(path) && r.methodMatcher(method) && r.headerMatcher(header)
}

func (r *Route) With(option Option) *Route {
	clone := r.Clone()
	option(clone)
	return clone
}

func (r *Route) Clone() *Route {
	return &Route{
		Path:          r.Path,
		PathMatcher:   r.PathMatcher,
		Method:        r.Method,
		MethodMatcher: r.MethodMatcher,
		HeaderMatcher: r.HeaderMatcher,
		Router: &Router{
			To:         r.Router.To,
			InstanceId: r.Router.InstanceId,
			Scheme:     r.Router.Scheme,
			Host:       r.Router.Host,
			Rewrite:    r.Router.Rewrite,
		},
		Filters:       r.Filters,
		pathMatcher:   r.pathMatcher,
		methodMatcher: r.methodMatcher,
		headerMatcher: r.headerMatcher,
		provider:      r.provider,
		reverseProxy:  r.reverseProxy,
	}
}

func (r *Route) Validate() error {
	if r.Path == "" {
		return errors.Errorf("path can not be empty in route %s", strutil.TryGetYamlStr(r))
	}
	if !strings.HasPrefix(r.Path, "/") {
		return errors.Errorf("path shoud has prefix / in route %s", strutil.TryGetYamlStr(r))
	}
	if err := r.genPathMatcher(); err != nil {
		return err
	}
	if err := r.genMethodMatcher(); err != nil {
		return err
	}
	if err := r.genHeaderMatcher(); err != nil {
		return err
	}
	if r.Router == nil {
		return errors.Errorf("router is not configurated in route %s", strutil.TryGetYamlStr(r))
	}
	if err := r.Router.validate(); err != nil {
		return err
	}
	for _, filter := range r.Filters {
		if _, ok := reverseproxy.GetFilterCreator(filter.Name); !ok {
			return errors.Errorf("no such filter creator named %s, do you import it ?", filter.Name)
		}
	}
	if r.provider == nil {
		r.provider = &provider.Provider{
			Name:       "__virtual__",
			InstanceId: "default",
			Host:       r.Router.Host,
			Scheme:     r.Router.Scheme,
			Metadata:   make(map[string]string),
		}
	}
	return nil
}

func (r *Route) GetProvider() *provider.Provider {
	return r.provider
}

func (r *Route) PathRegexExpr() string {
	_ = r.genPathMatcher()
	return r.PathMatcher
}

func (r *Route) genPathMatcher() error {
	if r.PathMatcher != "" {
		pathMatcher, err := regexp.Compile(r.PathMatcher)
		if err != nil {
			return errors.Wrapf(err, "Path %s is not a valid path or PathMatcher %s is not a valid regex expression",
				r.Path, r.PathMatcher)
		}
		r.pathMatcher = func(path string) bool {
			return pathMatcher.MatchString(path)
		}
		return nil
	}

	var p = r.Path
	for {
		expr, start, end, err := strutil.FirstCustomPlaceholder(p, "{", "}")
		if err != nil || start == end {
			break
		}
		p = strutil.Replace(p, `(?P<`+(expr)+`>[^/]+)`, start, end)
	}

	r.PathMatcher = `^` + p + `$`
	return r.genPathMatcher()
}

func (r *Route) genMethodMatcher() error {
	if r.Method == "" || r.Method == "*" || strings.EqualFold(r.Method, "any") {
		r.methodMatcher = func(string) bool { return true }
		return nil
	}
	switch r.MethodMatcher {
	case "contains", "in", "has":
		r.methodMatcher = func(method string) bool {
			return strings.Contains(strings.ToUpper(r.Method), strings.ToUpper(method))
		}
		return nil
	case "", "exact", "=", "==":
		r.methodMatcher = func(method string) bool {
			return strings.EqualFold(r.Method, method)
		}
		return nil
	default:
		return errors.Errorf("invalid method matcher exprestion %s", r.MethodMatcher)
	}
}

// genHeaderMatcher todo: not implement yet
func (r *Route) genHeaderMatcher() error {
	r.headerMatcher = func(header http.Header) bool {
		return true
	}
	return nil
}

func (r *Route) rewrite() string {
	if r.Router.Rewrite == "" {
		return r.Path
	}
	var rewrite = r.Router.Rewrite
	for {
		expr, start, end, err := strutil.FirstCustomExpression(rewrite, "${", "}", func(s string) bool {
			s = strings.TrimSpace(s)
			return strings.HasPrefix(s, "env.") || strings.HasPrefix(s, "provider.metadata.")
		})
		if err != nil || start == end {
			break
		}
		if strings.HasPrefix(expr, "env.") {
			rewrite = strutil.Replace(rewrite, os.Getenv(strings.TrimPrefix(expr, ".env")), start, end)
		} else {
			rewrite = strutil.Replace(rewrite, r.provider.Metadata[strings.TrimPrefix(expr, "provider.metadata.")], start, end)
		}
	}
	return rewrite
}

type Router struct {
	To         To     `json:"to" yaml:"to"`
	InstanceId string `json:"instanceId" yaml:"instanceId"`
	Scheme     string `json:"scheme" yaml:"scheme"`
	Host       string `json:"host" yaml:"host"`
	Rewrite    string `json:"rewrite" yaml:"rewrite"`
}

func (r *Router) validate() error {
	if r.To == "" {
		r.To = ToNotFound
	}
	if r.InstanceId == "" {
		r.InstanceId = "default"
	}
	if r.Scheme == "" {
		r.Scheme = "http"
	}
	switch r.Scheme {
	case "http", "https":
	default:
		return errors.Errorf("invlaid scheme %s, it must be one of http or https", r.Scheme)
	}
	if r.To == ToURL && r.Host == "" {
		return errors.Errorf("host can not be empty if route to %s", ToURL)
	}
	return nil
}

type To string

type Option func(route *Route)

func WithProvider(prov *provider.Provider) Option {
	return func(r *Route) {
		r.provider = prov
		if !strings.HasPrefix(string(r.Router.To), "__") || !strings.HasSuffix(string(r.Router.To), "__") {
			if r.Router.Scheme == "" && (prov.Scheme == "http" || prov.Scheme == "https") {
				r.Router.Scheme = prov.Scheme
			}
			r.Router.Host = prov.Host
		}
	}
}

func WithContext(ctx *reverseproxy.Context) Option {
	return func(route *Route) {
		route.reverseProxy.FilterCxt = ctx
	}
}
