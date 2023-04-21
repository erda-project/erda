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

package provider

import (
	"encoding/json"
)

const (
	ChatGPTv1 = "chatgpt/v1"
)

type Provider struct {
	Name        string `json:"name" yaml:"name"`
	Host        string `json:"host" yaml:"host"`
	Scheme      string `json:"scheme" yaml:"scheme"`
	Description string `json:"description" yaml:"description"`
	DocSite     string `json:"docSite" yaml:"docSite"`

	// appKey provided by ai-proxy, you can use an expression like ${ env.CHATGPT_APP_KEY }
	AppKey string `json:"appKey" yaml:"appKey"`
	// secretKey provided by ai-proxy, you can use an expression like ${ env.CHATGPT_APP_KEY }
	Organization string `json:"organization" yaml:"organization"`
	APIs         []*API `json:"apis" yaml:"apis"`
}

func (p *Provider) GetAppKey() string {
	return p.AppKey // todo: get from env expr
}

func (p *Provider) GetOrganization() string {
	return p.Organization // todo: get from env expr
}

func (p *Provider) FindAPI(name, path string) (*API, bool) {
	for _, api := range p.APIs {
		if api.Name == name {
			return api, true
		}
	}
	for _, api := range p.APIs {
		if api.MatchPath(path) {
			return api, true
		}
	}
	return nil, false
}

type API struct {
	Name    string          `json:"name" yaml:"name"`
	Path    string          `json:"path" yaml:"path"`
	Swagger json.RawMessage `json:"swagger" yaml:"swagger"`
}

func (api *API) MatchPath(path string) bool {
	// todo: not implement
	return true
}

type Providers []*Provider

func (p Providers) GetProvider(name string) (*Provider, bool) {
	for _, provider := range p {
		if provider.Name == name {
			return provider, true
		}
	}
	return nil, false
}
