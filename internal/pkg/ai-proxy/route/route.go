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
	"regexp"

	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	BackendTypeProvider BackendType = "provider"
	BackendTypeURL      BackendType = "url"

	ProtocolHTTP      Protocol = "http"
	ProtocolChatGPTv1 Protocol = "chatgpt/v1"

	ProviderChatGPTv1 Provider = "chatgpt/v1"
	ProviderAzure     Provider = "azure service"
)

type Routes []*Route

func (routes Routes) FindRoute(path, method string) (*Route, bool) {
	// todo: 应当改成树形数据结构来存储和查找 route, 不过在 route 数量有限的情形下影响不大
	for _, route := range routes {
		if route.Match(path, method) {
			return route, true
		}
	}
	return nil, false
}

type Route struct {
	Path    string           `json:"path" yaml:"path"`
	Method  string           `json:"method" yaml:"method"`
	Filters []*filter.Config `json:"filters" yaml:"filters"`

	matcher *regexp.Regexp
}

func (r *Route) Match(path, method string) bool {
	return r.MatchMethod(method) && r.MatchPath(path)
}

func (r *Route) MatchMethod(method string) bool {
	return strutil.Equal(r.Method, method)
}

func (r *Route) MatchPath(path string) bool {
	return r.getMatcher().MatchString(path)
}

func (r *Route) PathRegex() string {
	return r.getMatcher().String()
}

func (r *Route) getMatcher() *regexp.Regexp {
	var p = r.Path
	if r.matcher == nil {
		for i := 0; i < len(p); i++ {
			if p[i] == '{' {
				for j := i; i < len(p); j++ {
					if p[j] == '}' {
						p = p[:i] + `([^/.]+)` + p[j+1:]
						break
					}
				}
			}
		}
		r.matcher = regexp.MustCompile(`^` + p + `$`)
	}
	return r.matcher
}

type Backend struct {
	Type     BackendType `json:"type" yaml:"type"`
	Provider Provider    `json:"provider" yaml:"provider"`
	Name     string      `json:"name" yaml:"name"`
	Path     string      `json:"path" yaml:"path"`
}

type Protocol string

type BackendType string

type Provider string
