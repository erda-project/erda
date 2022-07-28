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
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/impl/kratos"
	"github.com/erda-project/erda/internal/core/user/impl/uc"
)

type config struct {
	OryEnabled bool `default:"false" file:"ORY_ENABLED" env:"ORY_ENABLED"`
}

type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register

	Kratos kratos.Interface
	Uc     uc.Interface

	userService common.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.OryEnabled {
		p.userService = p.Kratos
		p.Log.Info("use kratos as user")
	} else {
		p.userService = p.Uc
		p.Log.Info("use uc as user")
	}

	if p.Register != nil {
		pb.RegisterUserServiceImp(p.Register, p.userService)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.user.UserService" || ctx.Type() == pb.UserServiceServerType() || ctx.Type() == pb.UserServiceHandlerType():
		return p.userService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.user", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		ConfigFunc:           func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
