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

package discover

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type config struct {
	URLScheme string `file:"url_scheme" default:"http"`
	Endpoints map[string]string
	URLs      map[string]string
}

// +provider
type provider struct {
	Cfg *config
}

func (p *provider) Endpoint(service string) (string, error) {
	if addr, ok := p.Cfg.Endpoints[service]; ok {
		return addr, nil
	}
	return "", fmt.Errorf("not found endpoint %q", service)
}

func (p *provider) ServiceURL(service string) (string, error) {
	if url, ok := p.Cfg.URLs[service]; ok {
		return url, nil
	}
	if addr, ok := p.Cfg.Endpoints[service]; ok {
		return fmt.Sprintf("%s://%s", p.Cfg.URLScheme, addr), nil
	}
	return "", fmt.Errorf("not found url of service %q", service)
}

func init() {
	servicehub.Register("fixed-discover", &servicehub.Spec{
		Services:    []string{"discover"},
		Description: "discover all services",
		ConfigFunc: func() interface{} {
			return &config{
				Endpoints: make(map[string]string),
				URLs:      make(map[string]string),
			}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
