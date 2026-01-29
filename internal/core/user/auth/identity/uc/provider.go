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

package uc

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/user/identity/pb"
)

type Config struct {
	BackendHost string `file:"host"`
	ClientID    string `file:"client_id"`
	CookieName  string `file:"cookie_name"`
}

type provider struct {
	Register transport.Register `autowired:"service-register" required:"true"`
	Cfg      *Config
}

func (p *provider) Init(_ servicehub.Context) error {
	if p.Register != nil {
		pb.RegisterUserIdentityServiceImp(p.Register, p)
	}
	return nil
}

func init() {
	servicehub.Register("erda.core.user.identity.uc", &servicehub.Spec{
		Services:   pb.ServiceNames(),
		Types:      pb.Types(),
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
