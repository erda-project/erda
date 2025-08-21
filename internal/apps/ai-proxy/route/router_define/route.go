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

package router_define

import (
	"errors"
	"strings"

	"github.com/mohae/deepcopy"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define/path_matcher"
)

type Route struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`

	RequestFilters  []filter_define.FilterConfig `yaml:"request_filters,omitempty"`
	ResponseFilters []filter_define.FilterConfig `yaml:"response_filters,omitempty"`

	pathMatcher   *path_matcher.PathMatcher `yaml:"-"`
	methodMatcher func(method string) bool  `yaml:"-"`
}

func (r *Route) String() string {
	return r.Path + " " + r.Method
}

func (r *Route) InitMatchers() {
	if r.pathMatcher != nil {
		return
	}
	r.pathMatcher = path_matcher.NewPathMatcher(r.Path)

	r.methodMatcher = func(method string) bool {
		return strings.EqualFold(method, r.Method)
	}
}

// Implement router_define.RouteMatcher interface
func (r *Route) MethodMatcher(method string) bool {
	if r.methodMatcher == nil {
		r.InitMatchers()
	}
	return r.methodMatcher(method)
}

func (r *Route) PathMatcher() PathMatcher {
	if r.pathMatcher == nil {
		r.InitMatchers()
	}
	return r.pathMatcher
}

// Get original PathMatcher (for backward compatibility)
func (r *Route) GetPathMatcher() *path_matcher.PathMatcher {
	if r.pathMatcher == nil {
		r.InitMatchers()
	}
	return r.pathMatcher
}

func (r *Route) GetPath() string {
	return r.Path
}

func (r *Route) GetMethod() string {
	return r.Method
}

// ExpandRoute splits comma-separated paths and methods from a route configuration
// and returns a slice of individual routes with single path and method
func ExpandRoute(route *Route) ([]*Route, error) {
	var expandedRoutes []*Route

	paths := strings.Split(route.Path, ",")
	methods := strings.Split(route.Method, ",")

	for _, path := range paths {
		if path == "" {
			continue
		}
		for _, method := range methods {
			if method == "" {
				continue
			}
			if method == "*" {
				return nil, errors.New("method cannot be *")
			}
			newRoute := deepcopy.Copy(route).(*Route)
			newRoute.Path = path
			newRoute.Method = method
			expandedRoutes = append(expandedRoutes, newRoute)
		}
	}

	return expandedRoutes, nil
}
