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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	ToURL      To = "__url__"
	ToNotFound To = "__not_found__"
)

var (
	NotFoundRoute = &Route{
		Path:   "/",
		Router: &Router{To: ToNotFound},
	}
)

type Routes []*Route

func (routes Routes) FindRoute(path, method string, header http.Header) *Route {
	// todo: 应当改成树形数据结构来存储和查找 route, 不过在 route 数量有限的情形下影响不大
	for _, route := range routes {
		route := route.Clone()
		if route.Match(path, method, header) {
			return route
		}
	}
	return NotFoundRoute
}

type Route struct {
	Path          string                       `json:"path" yaml:"path"`
	PathMatcher   string                       `json:"pathMatcher" yaml:"pathMatcher"`
	Method        string                       `json:"method" yaml:"method"`
	MethodMatcher string                       `json:"methodMatcher" yaml:"methodMatcher"`
	HeaderMatcher json.RawMessage              `json:"headerMatcher" yaml:"headerMatcher"`
	Router        *Router                      `json:"router" yaml:"router"`
	Provider      *provider.Provider           `json:"provider" yaml:"provider"`
	Filters       []*reverseproxy.FilterConfig `json:"filters" yaml:"filters"`

	matchPath   func(path string) bool
	matchMethod func(method string) bool
	matchHeader func(header http.Header) bool
	vars        map[string]string
}

func (r *Route) HandlerWith(ctx context.Context, kvs ...any) http.HandlerFunc {
	// panic early
	if len(kvs)%2 != 0 {
		panic("kvs must be even")
	}

	// return "not found" early
	if r.IsNotFoundRoute() {
		return func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Server", "AI Service on Erda")
			http.Error(rw, string(ToNotFound), http.StatusNotFound)
		}
	}

	// to set logs.Logger and *provider.Provider into context.Context if not set
	var (
		l    logs.Logger
		prov *provider.Provider
	)
	for i := 0; i+1 < len(kvs); i++ {
		switch t := kvs[i+1].(type) {
		case logs.Logger:
			l = t.Sub(reflect.TypeOf(r).String())
		case *provider.Provider:
			prov = t
		}
		ctx = context.WithValue(ctx, kvs[i], kvs[i+1])
	}
	if l == nil {
		l = logrusx.New(logrusx.WithName(reflect.TypeOf(r).String()))
		ctx = context.WithValue(ctx, reverseproxy.LoggerCtxKey{}, l)
	}
	if prov == nil {
		prov = r.Provider
		ctx = context.WithValue(ctx, vars.CtxKeyProvider{}, prov)
	}

	// make reverseproxy.ReverseProxy and set filters
	var rp = &reverseproxy.ReverseProxy{
		Director: r.Director(ctx),
		Transport: &reverseproxy.TimerTransport{
			Logger: l,
			Inner: &reverseproxy.CurlPrinterTransport{
				Logger: l,
			},
		},
		FlushInterval: time.Millisecond * 100,
		BufferPool:    reverseproxy.DefaultBufferPool,
		Filters:       nil,
		Context:       ctx,
	}
	for _, filterConfig := range r.Filters {
		filter, err := reverseproxy.MustGetFilterCreator(filterConfig.Name)(filterConfig.Config)
		if err != nil {
			return func(rw http.ResponseWriter, req *http.Request) {
				l.Errorf("filter %s not found, err: %v", filterConfig.Name, err)
				http.Error(rw, fmt.Sprintf("filter %s not found: %v", filterConfig.Name, err), http.StatusNotImplemented)
			}
		}
		switch filter.(type) {
		case reverseproxy.RequestFilter, reverseproxy.ResponseFilter:
			rp.Filters = append(rp.Filters, reverseproxy.NamingFilter{Name: filterConfig.Name, Filter: filter})
		default:
			return func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, fmt.Sprintf("filter %s not found", filterConfig.Name), http.StatusInternalServerError)
			}
		}
	}

	return rp.ServeHTTP
}

func (r *Route) Match(path, method string, header http.Header) bool {
	return r.matchPath(path) && r.matchMethod(method) && r.matchHeader(header)
}

func (r *Route) Director(ctx context.Context) func(req *http.Request) {
	var prov = r.Provider
	if prov_, ok := ctx.Value(vars.CtxKeyProvider{}).(*provider.Provider); ok {
		prov = prov_
	}
	return func(req *http.Request) {
		switch {
		case r.Router.Scheme != "":
			req.URL.Scheme = r.Router.Scheme
		case prov == nil, prov.Scheme == "":
			req.URL.Scheme = "https"
		case prov.Scheme == "https", prov.Scheme == "http":
			req.URL.Scheme = prov.Scheme
		default:
			req.URL.Scheme = "https"
		}
		switch {
		case r.Router.Host != "":
			req.Host = r.Router.Host
		case prov == nil || prov.Host == "":
		default:
			req.Host = prov.GetHost()
		}
		req.URL.Host = req.Host
		req.Header.Set("Host", req.Host)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.URL.Path = r.RewritePath(prov.Metadata)
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
			return errors.Errorf("no such filter creator named %s, did you import it ?", filter.Name)
		}
	}

	return nil
}

func (r *Route) IsNotFoundRoute() bool {
	return r.Router != nil && strutil.Equal(r.Router.To, ToNotFound, true)
}

func (r *Route) PathRegexExpr() string {
	_ = r.genPathMatcher()
	return r.PathMatcher
}

func (r *Route) RewritePath(metadata map[string]string) string {
	if r.Router.Rewrite == "" {
		return r.Path
	}
	var rewrite = r.Router.Rewrite
	for {
		expr, start, end, err := strutil.FirstCustomExpression(rewrite, "${", "}", func(s string) bool {
			s = strings.TrimSpace(s)
			return strings.HasPrefix(s, "env.") || strings.HasPrefix(s, "provider.metadata.") || strings.HasPrefix(s, "path.")
		})
		if err != nil || start == end {
			break
		}
		if strings.HasPrefix(expr, "env.") {
			rewrite = strutil.Replace(rewrite, os.Getenv(strings.TrimPrefix(expr, "env.")), start, end)
		}
		if strings.HasPrefix(expr, "provider.metadata") && len(metadata) > 0 {
			rewrite = strutil.Replace(rewrite, metadata[strings.TrimPrefix(expr, "provider.metadata.")], start, end)
		}
		if strings.HasPrefix(expr, "path.") && len(r.vars) > 0 {
			rewrite = strutil.Replace(rewrite, r.vars[strings.TrimPrefix(expr, "path.")], start, end)
		}
	}
	return rewrite
}

func (r *Route) Clone() *Route {
	return &Route{
		Path:          r.Path,
		PathMatcher:   r.PathMatcher,
		Method:        r.Method,
		MethodMatcher: r.MethodMatcher,
		HeaderMatcher: r.HeaderMatcher,
		Router:        r.Router.Clone(),
		Provider:      r.Provider,
		Filters:       r.Filters,
		matchPath:     r.matchPath,
		matchMethod:   r.matchMethod,
		matchHeader:   r.matchHeader,
		vars:          r.vars,
	}
}

func (r *Route) genPathMatcher() error {
	if len(r.vars) == 0 {
		r.vars = make(map[string]string)
	}
	if r.PathMatcher != "" {
		pathMatcher, err := regexp.Compile(r.PathMatcher)
		if err != nil {
			return errors.Wrapf(err, "Path %s is not a valid path or PathMatcher %s is not a valid regex expression",
				r.Path, r.PathMatcher)
		}
		r.matchPath = func(path string) bool {
			if !pathMatcher.MatchString(path) {
				return false
			}
			matches := pathMatcher.FindAllStringSubmatch(path, -1)
			for _, match := range matches {
				for i, name := range pathMatcher.SubexpNames() {
					if i > 0 && i < len(match) {
						r.vars[name] = match[i]
					}
				}
			}
			return true
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
		r.matchMethod = func(string) bool { return true }
		return nil
	}
	switch r.MethodMatcher {
	case "contains", "in", "has":
		r.matchMethod = func(method string) bool {
			return strings.Contains(strings.ToUpper(r.Method), strings.ToUpper(method))
		}
		return nil
	case "", "exact", "=", "==":
		r.matchMethod = func(method string) bool {
			return strings.EqualFold(r.Method, method)
		}
		return nil
	default:
		return errors.Errorf("invalid method matcher exprestion %s", r.MethodMatcher)
	}
}

// genHeaderMatcher todo: not implement yet
func (r *Route) genHeaderMatcher() error {
	r.matchHeader = func(header http.Header) bool {
		return true
	}
	return nil
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
	switch r.Scheme {
	case "", "http", "https":
	default:
		return errors.Errorf("invlaid scheme %s, it must be one of http or https", r.Scheme)
	}
	if r.To == ToURL && r.Host == "" {
		return errors.Errorf("host can not be empty if route to %s", ToURL)
	}
	return nil
}

func (r *Router) Clone() *Router {
	return &Router{
		To:         r.To,
		InstanceId: r.InstanceId,
		Scheme:     r.Scheme,
		Host:       r.Host,
		Rewrite:    r.Rewrite,
	}
}

type To string
