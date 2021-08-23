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

package services

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/runtime"
	discover "github.com/erda-project/erda/providers/service-discover"
)

// +provider
type provider struct {
	Log      logs.Logger
	Router   transhttp.Router   `autowired:"openapi-ng"`
	Discover discover.Interface `autowired:"discover"`
	err      error
}

func (p *provider) Init(ctx servicehub.Context) error {
	RegisterAPIs(p.Add)
	RegisterCustomAPIs(func(method, path string, handler http.HandlerFunc) {
		p.Router.Add(method, path, transhttp.HandlerFunc(handler))
	})
	return p.err
}

func (p *provider) Add(method, path, backendPath, service string) {
	if p.err != nil {
		return
	}
	srvURL, err := p.Discover.ServiceURL(service)
	if err != nil {
		p.err = fmt.Errorf("fail to discover url of service %q: %s", service, err)
		return
	}
	pubMatcher, err := runtime.Compile(path)
	if err != nil {
		p.err = fmt.Errorf("invalid path %q of service %q: %s", path, service, err)
		return
	}
	backMatcher, err := runtime.Compile(backendPath)
	if err != nil {
		p.err = fmt.Errorf("invalid backend-path %q of service %q: %s", backendPath, service, err)
		return
	}
	var director func(req *http.Request)
	if backMatcher.IsStatic() {
		director, err = p.staticDirector(srvURL, backendPath)
	} else {
		director, err = p.paramsDirector(srvURL, pubMatcher, backMatcher)
	}
	if err != nil {
		p.err = err
		return
	}
	rp := &httputil.ReverseProxy{Director: func(req *http.Request) {
		director(req)
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}}
	p.Router.Add(method, path, rp.ServeHTTP)
}

func (p *provider) staticDirector(srvURL, backendPath string) (func(req *http.Request), error) {
	srvURL = strings.TrimRight(srvURL, "/") + "/" + strings.TrimLeft(backendPath, "/")
	target, err := url.Parse(srvURL)
	if err != nil {
		return nil, fmt.Errorf("fail to parse service url: %s", err)
	}
	path, rawpath := target.Path, target.EscapedPath()
	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = path, rawpath
		p.Log.Debugf("proxy %s %s -> %s", req.Method, rawpath, req.URL)
	}, nil
}

func (p *provider) paramsDirector(srvURL string, pmatcher, bmatcher runtime.Matcher) (func(req *http.Request), error) {
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
	target, err := url.Parse(srvURL)
	if err != nil {
		return nil, fmt.Errorf("fail to parse service url: %s", err)
	}
	segments := buildPathToSegments(bmatcher.Pattern())
	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		params, _ := pmatcher.Match(req.URL.Path)
		sb := strings.Builder{}
		for _, seg := range segments {
			if seg.typ == pathStatic {
				sb.WriteString(seg.name)
			} else {
				sb.WriteString(params[seg.name])
			}
		}
		rawpath := req.URL.Path
		req.URL.Path, req.URL.RawPath = sb.String(), ""
		req.URL.RawPath = req.URL.EscapedPath()
		p.Log.Debugf("proxy %s %s -> %s", req.Method, rawpath, req.URL)
	}, nil
}

const pathStatic, pathField int8 = 0, 1

type pathSegment struct {
	typ  int8
	name string
}

func (s *pathSegment) String() string { return fmt.Sprint(*s) }

func buildPathToSegments(path string) (segs []*pathSegment) {
	chars := []rune(path)
	start, i, n := 0, 0, len(chars)
	for ; i < n; i++ {
		if chars[i] == '{' {
			if start < i {
				segs = append(segs, &pathSegment{
					typ:  pathStatic,
					name: string(chars[start:i]),
				})
			}
			i++
			for begin := i; i < n; i++ {
				switch chars[i] {
				case '}':
					segs = append(segs, &pathSegment{
						typ:  pathField,
						name: string(chars[begin:i]),
					})
					start = i + 1
					break
				case '=':
					segs = append(segs, &pathSegment{
						typ:  pathField,
						name: string(chars[begin:i]),
					})
					for ; i < n && chars[i] != '}'; i++ {
					}
					start = i + 1
					break
				}
			}
		}
	}
	if start < n {
		segs = append(segs, &pathSegment{
			typ:  pathStatic,
			name: string(chars[start:]),
		})
	}
	return
}

func init() {
	servicehub.Register("openapi-services", &servicehub.Spec{
		Services: []string{"openapi-services"},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
