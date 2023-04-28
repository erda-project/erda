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
	"os"
	"strings"

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

	Metadata map[string]string `json:"metadata" yaml:"metadata"`
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

type Providers []*Provider

func (p Providers) FindProvider(name, instanceId string) (*Provider, bool) {
	for _, provider := range p {
		if provider.Name == name && provider.InstanceId == instanceId {
			return provider, true
		}
	}
	return nil, false
}
