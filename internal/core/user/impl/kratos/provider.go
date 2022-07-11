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

package kratos

import (
	"reflect"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/user/common"
)

type Config struct {
	OryKratosPrivateAddr string `default:"kratos-admin" file:"ORY_KRATOS_ADMIN_ADDR" env:"ORY_KRATOS_ADMIN_ADDR"`
}

type provider struct {
	Cfg *Config
	Log logs.Logger
	DB  *gorm.DB `autowired:"mysql-client"`

	baseURL string
}

type IdentityService struct {
	Provider *provider
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.baseURL = p.Cfg.OryKratosPrivateAddr
	return nil
}

type Interface common.Interface

func init() {
	servicehub.Register("erda.core.user.kratos", &servicehub.Spec{
		Services:   []string{"erda.core.user.kratos"},
		Types:      []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &Config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
