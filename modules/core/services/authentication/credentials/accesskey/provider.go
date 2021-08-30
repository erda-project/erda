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

package accesskey

import (
	"github.com/jinzhu/gorm"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct {
}

// +provider
type provider struct {
	Cfg              *config
	Log              logs.Logger
	Register         transport.Register
	accessKeyService *accessKeyService
	dao              Dao
	Perm             perm.Interface `autowired:"permission"`
	DB               *gorm.DB       `autowired:"mysql-client"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.dao = &dao{db: p.DB}
	p.accessKeyService = &accessKeyService{p}
	if p.Register != nil {
		pb.RegisterAccessKeyServiceImp(p.Register, p.accessKeyService, apis.Options(), p.Perm.Check(
			perm.NoPermMethod(pb.AccessKeyServiceServer.QueryAccessKeys),
			perm.NoPermMethod(pb.AccessKeyServiceClient.GetAccessKey),
			perm.NoPermMethod(pb.AccessKeyServiceClient.CreateAccessKey),
			perm.NoPermMethod(pb.AccessKeyServiceClient.UpdateAccessKey),
			perm.NoPermMethod(pb.AccessKeyServiceClient.DeleteAccessKey),
		))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.services.authentication.credentials.accesskey.AccessKeyService" || ctx.Type() == pb.AccessKeyServiceServerType() || ctx.Type() == pb.AccessKeyServiceHandlerType():
		return p.accessKeyService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.services.authentication.credentials.accesskey", &servicehub.Spec{
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
