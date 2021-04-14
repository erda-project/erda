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

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/pkg/discover"
)

// Interface .
type Interface interface {
	Endpoint(service string) (string, error)
	ServiceURL(service string) (string, error)
}

type config struct{}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
}

func (p *provider) Endpoint(name string) (string, error) {
	return discover.GetEndpoint(name)
}

func (p *provider) ServiceURL(service string) (string, error) {
	endpoint, err := p.Endpoint(service)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s", endpoint), nil
}

func init() {
	servicehub.Register("erda-discover", &servicehub.Spec{
		Services:    []string{"discover"},
		Description: "discover all services",
		ConfigFunc:  func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
