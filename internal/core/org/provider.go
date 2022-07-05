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

package org

import (
	"reflect"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/redis"
	"github.com/erda-project/erda-proto-go/core/org/pb"
	tokenpb "github.com/erda-project/erda-proto-go/core/token/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/internal/core/legacy/services/member"
	"github.com/erda-project/erda/internal/core/legacy/services/permission"
	"github.com/erda-project/erda/internal/core/org/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/ucauth"
)

type config struct {
	// Allow people who are not admin to create org
	CreateOrgEnabled bool   `default:"false" env:"CREATE_ORG_ENABLED"`
	UIDomain         string `env:"UI_PUBLIC_ADDR"`
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register
	DB       *gorm.DB `autowired:"mysql-client" optional:"true"`
	RedisCli redis.Interface
	bdl      *bundle.Bundle
	dbClient *db.DBClient

	member     *member.Member
	uc         *ucauth.UCClient
	permission *permission.Permission

	TokenService tokenpb.TokenServiceServer
	Client       pb.OrgServiceServer
}

func (p *provider) WithMember(member *member.Member) {
	p.member = member
}

func (p *provider) WithUc(uc *ucauth.UCClient) {
	p.uc = uc
}

func (p *provider) WithPermission(permission *permission.Permission) {
	p.permission = permission
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithCoreServices(), bundle.WithClusterManager(), bundle.WithOrchestrator())
	p.dbClient = &db.DBClient{DBClient: &dao.DBClient{DB: p.DB}}
	if p.Client == nil && p.Register != nil {
		pb.RegisterOrgServiceImp(p.Register, p, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Type() == reflect.TypeOf((*ClientInterface)(nil)).Elem():
		return &WrapClient{p.Client}
	case ctx.Service() == "erda.core.org.OrgService" || ctx.Type() == pb.OrgServiceServerType() || ctx.Type() == pb.OrgServiceHandlerType():
		return p
	}
	return p
}

func init() {
	servicehub.Register("erda.core.org", &servicehub.Spec{
		Types:                []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem(), reflect.TypeOf((*ClientInterface)(nil)).Elem()},
		OptionalDependencies: []string{"service-register"},
		Description:          "org",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
