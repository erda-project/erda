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

package user

import (
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/impl/kratos"
	"github.com/erda-project/erda/internal/core/user/impl/uc"
)

type config struct {
	OryEnabled bool `default:"false" file:"ORY_ENABLED" env:"ORY_ENABLED"`
}

type provider struct {
	Cfg *config
	Log logs.Logger

	Kratos kratos.Interface
	Uc     uc.Interface

	common.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.OryEnabled {
		p.Interface = p.Kratos
		p.Log.Info("use kratos as user")
	} else {
		p.Interface = p.Uc
		p.Log.Info("use uc as user")
	}
	return nil
}

type Interface common.Interface

func init() {
	servicehub.Register("erda.core.user", &servicehub.Spec{
		Services:   []string{"erda.core.user"},
		Types:      []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
