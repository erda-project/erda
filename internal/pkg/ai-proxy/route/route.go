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

import "github.com/erda-project/erda/internal/pkg/ai-proxy/filter"

const (
	BackendTypeProvider BackendType = "provider"
	BackendTypeURL      BackendType = "url"

	ProtocolHTTP      Protocol = "http"
	ProtocolChatGPTv1 Protocol = "chatgpt/v1"

	ProviderChatGPTv1 Provider = "chatgpt/v1"
	ProviderAzure     Provider = "azure service"
)

type Routes []*Route

type Route struct {
	Path     string           `json:"path" yaml:"path"`
	Protocol Protocol         `json:"protocol" yaml:"protocol"`
	Methods  map[string]any   `json:"methods" yaml:"methods"`
	Filters  []*filter.Config `json:"filters" yaml:"filters"`
}

func (r *Route) Match(path, method string) bool {
	if _, ok := r.Methods[method]; !ok {
		return false
	}
	return r.Path == path // todo: 需要实现匹配方法
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
