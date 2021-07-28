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

package dingtalktest

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/admin/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg                 *config
	Log                 logs.Logger
	dingTalkTestService *dingTalkTestService
	Register            transport.Register `autowired:"service-register" optional:"true"`
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.dingTalkTestService = &dingTalkTestService{
		Log: p.Log,
		bdl: bundle.New(bundle.WithEventBox()),
	}
	if p.Register != nil {
		pb.RegisterDingTalkTestServiceImp(p.Register, p.dingTalkTestService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.admin.DingTalkTestService" || ctx.Type() == pb.DingTalkTestServiceServerType() || ctx.Type() == pb.DingTalkTestServiceHandlerType():
		return p.dingTalkTestService
	}
	return p
}

func init() {
	servicehub.Register("erda.admin.dingtalk-test", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
