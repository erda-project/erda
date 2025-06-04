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
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
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

func (routes Routes) FindRoute(req *http.Request) *Route {
	// todo: 应当改成树形数据结构来存储和查找 route, 不过在 route 数量有限的情形下影响不大
	for _, route := range routes {
		if clone := route.Clone(); clone.Match(req.URL.Path, req.Method) {
			return clone
		}
	}
	return NotFoundRoute
}

func (routes Routes) Validate() error {
	for _, item := range routes {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Route struct {
	Path      string                       `json:"path" yaml:"path"`
	PathRegex string                       `json:"pathRegex" yaml:"pathRegex"`
	Method    string                       `json:"method" yaml:"method"`
	Router    *Router                      `json:"router" yaml:"router"`
	Filters   []*reverseproxy.FilterConfig `json:"filters" yaml:"filters"`

	PathMatcher   *PathMatcher             `json:"-" yaml:"-"`
	MethodMatcher func(method string) bool `json:"-" yaml:"-"`
}

func (r *Route) HandlerWith(ctx context.Context, kvs ...any) http.HandlerFunc {
	// panic early
	if len(kvs)%2 != 0 {
		panic("kvs must be even")
	}

	// return "not found" early
	if r.IsNotFoundRoute() {
		return func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Server", "ai-proxy")
			http.Error(rw, string(ToNotFound), http.StatusNotFound)
		}
	}

	// set route path vars to context
	ctx = context.WithValue(ctx, reverseproxy.CtxKeyPathMatcher{}, r.PathMatcher)
	ctx = context.WithValue(ctx, reverseproxy.CtxKeyPath{}, r.Path)

	// set default logger to context
	var l = logrusx.New(logrusx.WithName(reflect.TypeOf(r).String()))
	ctx = context.WithValue(ctx, reverseproxy.LoggerCtxKey{}, l)

	// set key-values to context
	for i := 0; i+1 < len(kvs); i += 2 {
		if !IsComparable(kvs[i]) {
			l.Fatalf("the key [%d] is not comparable: %T", i, kvs[i])
		}
		ctx = context.WithValue(ctx, kvs[i], kvs[i+1])
	}
	// reload the logger
	l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)

	// make reverseproxy.ReverseProxy and set filters
	var rp = &reverseproxy.ReverseProxy{
		Director: r.Director(ctx),
		Transport: &reverseproxy.TimerTransport{
			Logger: l,
			Inner: &reverseproxy.CurlPrinterTransport{
				Logger: l,
			},
		},
		ModifyResponse: func() func(response *http.Response) (err error) {
			return ctx.Value(reverseproxy.CtxKeyModifyResponse{}).(func(*http.Response) error)
		}(),
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

func (r *Route) Match(path, method string) bool {
	return r.PathMatcher.Match(path) && r.MethodMatcher(method)
}

func (r *Route) Director(ctx context.Context) func(req *http.Request) {
	return func(req *http.Request) {
		// filters 可能会传递 directors, 如果传递了, 要依次执行
		if m, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map); ok && m != nil {
			if value, ok := m.Load(reverseproxy.MapKeyDirectors{}); ok && value != nil {
				if directors, ok := value.([]func(*http.Request)); ok {
					for _, director := range directors {
						director(req)
					}
				}
			}
		}
		// 如果配置中定义了 Router, 那么使用 Router 来 direct request
		if r.Router != nil && r.PathMatcher != nil {
			r.Router.Director(ctx, r.PathMatcher.Values)(req)
		}
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
	if r.Router != nil {
		if err := r.Router.validate(); err != nil {
			return err
		}
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
	return r.PathRegex
}

func (r *Route) Clone() *Route {
	route := &Route{
		Path:          r.Path,
		PathRegex:     r.PathRegex,
		Method:        r.Method,
		Router:        r.Router,
		Filters:       r.Filters,
		PathMatcher:   r.PathMatcher,
		MethodMatcher: r.MethodMatcher,
	}
	if route.Router != nil {
		route.Router = route.Router.Clone()
	}
	return route
}

// OverridePathRegex generates the PathRegex and overwrite it
func (r *Route) OverridePathRegex() {
	var p = r.Path
	for {
		expr, start, end, err := strutil.FirstCustomPlaceholder(p, "{", "}")
		if err != nil || start == end {
			break
		}
		p = strutil.Replace(p, `(?P<`+(expr)+`>[^/]+)`, start, end)
	}
	r.PathRegex = `^` + p + `$`
}

func (r *Route) genPathMatcher() (err error) {
	// if there is no PathRegex, generate it
	if r.PathRegex == "" {
		r.OverridePathRegex()
	}
	r.PathMatcher, err = NewPathMatcher(r.PathRegex)
	return err
}

func (r *Route) genMethodMatcher() error {
	if r.Method == "" || r.Method == "*" || strings.EqualFold(r.Method, "any") {
		r.MethodMatcher = func(string) bool { return true }
		return nil
	}
	r.MethodMatcher = func(method string) bool {
		return strings.EqualFold(r.Method, method)
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

func (r *Router) Director(ctx context.Context, metadata map[string]string) func(r *http.Request) {
	return func(req *http.Request) {
		// todo: support to direct with r.To and r.InstanceId

		if r.Scheme == "http" || r.Scheme == "https" {
			req.URL.Scheme = r.Scheme
		}
		if r.Host != "" {
			req.Host = r.Host
		}
		req.URL.Host = req.Host
		req.Header.Set("Host", req.Host)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.URL.Path = r.RewritePath(req.URL.Path, metadata)
	}
}

func (r *Router) RewritePath(path string, metadata map[string]string) string {
	if r.Rewrite == "" {
		return path
	}
	var rewrite = r.Rewrite
	for {
		expr, start, end, err := strutil.FirstCustomExpression(rewrite, "${", "}", func(s string) bool {
			s = strings.TrimSpace(s)
			return strings.HasPrefix(s, "env.") || strings.HasPrefix(s, "path.")
		})
		if err != nil || start == end {
			break
		}
		if strings.HasPrefix(expr, "env.") {
			rewrite = strutil.Replace(rewrite, os.Getenv(strings.TrimPrefix(expr, "env.")), start, end)
		}
		if strings.HasPrefix(expr, "path.") {
			rewrite = strutil.Replace(rewrite, metadata[strings.TrimPrefix(expr, "path.")], start, end)
		}
	}
	return rewrite
}

func (r *Router) Clone() *Router {
	if r == nil {
		return nil
	}
	return &Router{
		To:         r.To,
		InstanceId: r.InstanceId,
		Scheme:     r.Scheme,
		Host:       r.Host,
		Rewrite:    r.Rewrite,
	}
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

type To string

func NewPathMatcher(expr string) (*PathMatcher, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	var pm = PathMatcher{
		Values: make(map[string]string),
		mu:     new(sync.Mutex),
	}
	pm.match = func(path string) bool {
		if !re.MatchString(path) {
			return false
		}
		matches := re.FindAllStringSubmatch(path, -1)
		for _, subs := range matches {
			for i, name := range re.SubexpNames() {
				if i > 0 && i < len(subs) {
					pm.SetValue(name, subs[i])
				}
			}
		}
		return true
	}
	return &pm, nil
}

type PathMatcher struct {
	Values map[string]string

	match func(path string) bool
	mu    *sync.Mutex
}

func (p *PathMatcher) Match(path string) bool {
	return p.match(path)
}

func (p *PathMatcher) SetValue(key, value string) {
	p.mu.Lock()
	p.Values[key] = value
	p.mu.Unlock()
}

func IsComparable(v any) bool {
	rv := reflect.ValueOf(v)
	t := rv.Type()
	switch t.Kind() {
	case reflect.Array:
		return IsComparable(reflect.Zero(t.Elem()).Interface())
	default:
		return t.Comparable()
	}
}
