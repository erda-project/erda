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

package adapter

import (
	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/msp/apm/adapter/pb"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg            *config
	Log            logs.Logger
	Register       transport.Register
	adapterService *adapterService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.adapterService = &adapterService{p}
	if p.Register != nil {
		pb.RegisterAdapterServiceImp(p.Register, p.adapterService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.apm.adapter.AdapterService" || ctx.Type() == pb.AdapterServiceServerType() || ctx.Type() == pb.AdapterServiceHandlerType():
		return p.adapterService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.apm.adapter", &servicehub.Spec{
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
