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
	"github.com/erda-project/erda/pkg/discover"
)

type config struct {
	URLScheme string `file:"url_scheme" default:"http"`
}

// +provider
type provider struct {
	Cfg *config
}

func (p *provider) Endpoint(service string) (string, error) {
	return discover.GetEndpoint(service)
}

func (p *provider) ServiceURL(service string) (string, error) {
	endpoint, err := p.Endpoint(service)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s", p.Cfg.URLScheme, endpoint), nil
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
