// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package openapi

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/openapi-ng/api"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors"
	discover "github.com/erda-project/erda/providers/service-discover"
)

var _ Interface = (*router)(nil)

type router struct {
	name         string
	log          logs.Logger
	discover     discover.Interface `autowired:"discover"`
	interceptors interceptors.Interceptors
	http         httpserver.Router
	addError     func(error)
}

func (r *router) Add(method, path string, handler transhttp.HandlerFunc) {
	for i := len(r.interceptors) - 1; i >= 0; i-- {
		handler = transhttp.HandlerFunc(r.interceptors[i].Wrapper(http.HandlerFunc(handler)))
	}
	r.http.Add(method, path, handler, httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs))
}

func (r *router) AddAPI(spec *api.Spec) {
	err := r.addAPI(spec)
	if err != nil {
		r.addError(err)
	}
}

func (r *router) addAPI(spec *api.Spec) error {
	if len(spec.ServiceURL) <= 0 && len(spec.Service) > 0 {
		srvURL, err := r.discover.ServiceURL(spec.Service)
		if err != nil {
			return fmt.Errorf("fail to discover url for service %q: %s", spec.Service, err)
		}
		spec.ServiceURL = srvURL
	}
	apiCtx, err := api.NewContext(spec)
	if err != nil {
		return err
	}
	handler, err := r.getHandler(apiCtx)
	if err != nil {
		return err
	}
	for i := len(r.interceptors) - 1; i >= 0; i-- {
		handler = r.interceptors[i].Wrapper(handler)
	}
	apiHandler := func(rw http.ResponseWriter, r *http.Request) {
		r = r.WithContext(api.WithContext(r.Context(), apiCtx))
		handler(rw, r)
	}
	r.http.Add(spec.Method, spec.Path, apiHandler, httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs))
	return nil
}

func (r *router) getHandler(ctx api.Context) (_ http.HandlerFunc, err error) {
	spec := ctx.Spec()
	if spec.Handler != nil {
		return spec.Handler, nil
	}
	var director func(req *http.Request)
	if ctx.BackendMatcher().IsStatic() {
		director, err = r.staticDirector(ctx)
	} else {
		director, err = r.paramsDirector(ctx)
	}
	if err != nil {
		return nil, err
	}
	rp := &httputil.ReverseProxy{Director: func(req *http.Request) {
		director(req)
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}}
	return rp.ServeHTTP, nil
}

func (r *router) staticDirector(ctx api.Context) (func(req *http.Request), error) {
	backendURL := strings.TrimRight(ctx.ServiceURL().String(), "/") + "/" + strings.TrimLeft(ctx.Spec().BackendPath, "/")
	target, err := url.Parse(backendURL)
	if err != nil {
		return nil, fmt.Errorf("fail to parse backend url: %s", err)
	}
	path, rawpath := target.Path, target.EscapedPath()
	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = path, rawpath
		r.log.Debugf("proxy %s %s -> %s", req.Method, rawpath, req.URL)
	}, nil
}

func (r *router) paramsDirector(ctx api.Context) (func(req *http.Request), error) {
	pmatcher, bmatcher := ctx.Matcher(), ctx.BackendMatcher()
	if pmatcher.IsStatic() {
		return nil, fmt.Errorf("backend-path:%s has parameters, but publish-path:%s is static", bmatcher.Pattern(), pmatcher.Pattern())
	}
	fields := bmatcher.Fields()
	for _, field := range fields {
		var find bool
		for _, key := range pmatcher.Fields() {
			if field == key {
				find = true
				break
			}
		}
		if !find {
			return nil, fmt.Errorf("backend-path:%s has parameter %q, but not present in publish-path:%s", bmatcher.Pattern(), field, pmatcher.Pattern())
		}
	}
	target := ctx.ServiceURL()
	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		vars, _ := pmatcher.Match(req.URL.Path)
		rawpath := req.URL.Path
		req.URL.Path, req.URL.RawPath = ctx.BackendPath(vars), ""
		req.URL.RawPath = req.URL.EscapedPath()
		r.log.Debugf("proxy %s %s -> %s", req.Method, rawpath, req.URL)
	}, nil
}
