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
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/pkg/strutil"
)

const (
	ChatGPTv1 = "chatgpt/v1"
)

type Provider struct {
	Name        string `json:"name" yaml:"name"`
	InstanceId  string `json:"instanceId" yaml:"instanceId"`
	Host        string `json:"host" yaml:"host"`
	Scheme      string `json:"scheme" yaml:"scheme"`
	Description string `json:"description" yaml:"description"`
	DocSite     string `json:"docSite" yaml:"docSite"`

	// appKey provided by ai-proxy, you can use an expression like ${ env.CHATGPT_APP_KEY }
	AppKey string `json:"appKey" yaml:"appKey"`

	// secretKey provided by ai-proxy, you can use an expression like ${ env.CHATGPT_APP_KEY }
	Organization string `json:"organization" yaml:"organization"`

	Openapi  json.RawMessage   `json:"openapi" json:"openapi"`
	Metadata map[string]string `json:"metadata" yaml:"metadata"`
	Swagger  *openapi3.Swagger `json:"-" yaml:"-"`
}

func (p *Provider) GetHost() string {
	return p.getRendered(p.Host)
}

func (p *Provider) GetAppKey() string {
	return p.getRendered(p.AppKey)
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

func (p *Provider) FindAPI(operationId, path, method string) (*openapi3.Operation, bool) {
	if operationId != "" {
		return p.findAPIByOperationId(operationId)
	}
	return p.findAPIByPathMethod(path, method)
}

func (p *Provider) getRendered(s string) string {
	for {
		expr, start, end, err := strutil.FirstCustomExpression(s, "${", "}", func(s string) bool {
			s = strings.TrimSpace(s)
			return strings.HasPrefix(s, "env.") || strings.HasPrefix(s, "metadata.")
		})
		if err != nil || start == end {
			break
		}
		if strings.HasPrefix(expr, "env.") {
			s = strutil.Replace(s, os.Getenv(strings.TrimPrefix(expr, "env.")), start, end)
		} else if strings.HasPrefix(expr, "metadata.") {
			s = strutil.Replace(s, p.Metadata[strings.TrimPrefix(expr, "metadata.")], start, end)
		}
	}
	return s
}

func (p *Provider) findAPIByOperationId(operationId string) (*openapi3.Operation, bool) {
	if p.Swagger == nil || len(p.Swagger.Paths) == 0 {
		return nil, false
	}
	for _, item := range p.Swagger.Paths {
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
			if operation != nil && operation.OperationID == operationId {
				return operation, true
			}
		}
	}
	return nil, false
}

func (p *Provider) findAPIByPathMethod(path, method string) (*openapi3.Operation, bool) {
	panic(fmt.Sprintf("%T.findAPIByPathMethod not implement", p))
}

type Providers []*Provider

func (p Providers) FindProvider(name, instanceId string) (*Provider, bool) {
	for _, provider := range p {
		if provider.Name == name && provider.InstanceId == instanceId {
			return provider, true
		}
	}
	return nil, false
}

func MatchPath(pattern, path string) bool {
	return true // todo: not implement
}
