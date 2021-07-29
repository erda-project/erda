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

package image

import (
	"github.com/jinzhu/gorm"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/core/dicehub/image/pb"
	db "github.com/erda-project/erda/modules/dicehub/image/db"
)

type config struct {
}

// +provider
type provider struct {
	Cfg          *config
	Log          logs.Logger
	Register     transport.Register `autowired:"service-register" required:"true"`
	DB           *gorm.DB           `autowired:"mysql-client"`
	imageService *imageService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.imageService = &imageService{
		p:  p,
		db: &db.ImageConfigDB{DB: p.DB},
	}
	if p.Register != nil {
		pb.RegisterImageServiceImp(p.Register, p.imageService)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.dicehub.image.ImageService" || ctx.Type() == pb.ImageServiceServerType() || ctx.Type() == pb.ImageServiceHandlerType():
		return p.imageService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.dicehub.image", &servicehub.Spec{
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
