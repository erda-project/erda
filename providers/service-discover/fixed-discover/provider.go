// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
