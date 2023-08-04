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

package external_openapi

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	common "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes"
)

var (
	name = "erda.openapi-ng.external-openapi"
	spec = servicehub.Spec{
		Summary:     "external apis expose in erda openapi",
		Description: "external apis expose in erda openapi",
		ConfigFunc:  func() any { return new(Config) },
		Creator:     func() servicehub.Provider { return new(provider) },
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	Log     logs.Logger
	Config  *Config
	Openapi routes.Register `autowired:"openapi-dynamic-register.client"`
}

func (p *provider) Init(_ servicehub.Context) error {
	if p.Config == nil {
		return errors.New("config is invalid")
	}
	for _, api := range p.Config.APIs {
		proxy := &routes.APIProxy{
			Method:      strings.ToUpper(api.Method),
			Path:        api.Path,
			ServiceURL:  p.Config.Service.URL,
			BackendPath: api.BackendPath,
			Auth:        api.Auth,
		}
		if proxy.BackendPath == "" {
			proxy.BackendPath = proxy.Path
		}
		if proxy.Auth == nil {
			proxy.Auth = p.Config.Service.Auth
		}
		p.Log.Infof("register external API to openapi, %s %s -> %s%s\n", proxy.Method, proxy.Path, proxy.ServiceURL, proxy.BackendPath)
		if err := p.Openapi.Register(proxy); err != nil {
			return err
		}
	}
	return nil
}

type Config struct {
	Service Service `json:"service" yaml:"service"`
	APIs    []API   `json:"apis" yaml:"apis"`
}

type Service struct {
	Name string          `json:"name" yaml:"name"`
	URL  string          `json:"url" yaml:"url"`
	Auth *common.APIAuth `json:"auth" yaml:"auth"`
}

type API struct {
	Name        string          `json:"name" yaml:"name"`
	Method      string          `json:"method" yaml:"method"`
	Path        string          `json:"path" yaml:"path"`
	BackendPath string          `json:"backendPath" yaml:"backendPath"`
	Auth        *common.APIAuth `json:"auth" yaml:"auth"`
}
