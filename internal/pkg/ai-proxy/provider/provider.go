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

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/yaml"
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
	Organization string            `json:"organization" yaml:"organization"`
	Openapi      json.RawMessage   `json:"openapi" json:"openapi"`
	Swagger      *openapi3.Swagger `json:"-" yaml:"-"`
}

func (p *Provider) GetAppKey() string {
	return p.AppKey // todo: get from env expr
}

func (p *Provider) GetOrganization() string {
	return p.Organization // todo: get from env expr
}

func (p *Provider) LoadOpenapiSpec() error {
	var m = make(map[string]string)
	if err := yaml.Unmarshal(p.Openapi, &m); err == nil {
		if filename, ok := m["$ref"]; ok && filename != "" {
			swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(filename)
			if err != nil {
				return err
			}
			p.Swagger = swagger
			return nil
		}
	}
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(p.Openapi)
	if err != nil {
		return err
	}
	p.Swagger = swagger
	return nil
}

func (p *Provider) FindAPI(name, path string) bool {
	for pth, item := range p.Swagger.Paths {
		if path == "" {
			for _, operation := range []*openapi3.Operation{
				item.Connect,
				item.Delete,
				item.Get,
				item.Head,
				item.Options,
				item.Patch,
				item.Post,
				item.Put,
				item.Trace,
			} {
				if operation != nil && name == operation.OperationID {
					return true
				}
			}
		}
		if MatchPath(pth, path) {
			return true
		}
	}
	return false
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

func MatchPath(pattern, path string) bool {
	return true // todo: not implement
}
