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

package contribution

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/dop/contribution/pb"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type config struct {
}

type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register `autowired:"service-register" required:"true"`
	DB       *gorm.DB           `autowired:"mysql-client"`
	I18n     i18n.Translator    `autowired:"i18n" translator:"contribution"`
	Perm     perm.Interface     `autowired:"permission"`

	contributionService pb.ContributionServiceServer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.contributionService = &contributionService{
		db: &dao.DBClient{
			DBEngine: &dbengine.DBEngine{
				DB: p.DB,
			},
		},
		logger: p.Log,
		i18n:   p.I18n,
	}

	if p.Register != nil {
		pb.RegisterContributionServiceImp(p.Register, p.contributionService, apis.Options(), p.Perm.Check(
			perm.Method(p.contributionService.GetActiveRank, perm.ScopeOrg, perm.ScopeOrg, perm.ActionGet, func(ctx context.Context, req interface{}) (string, error) {
				return req.(*pb.GetActiveRankRequest).OrgID, nil
			}),
			perm.Method(p.contributionService.GetPersonalContribution, perm.ScopeOrg, perm.ScopeOrg, perm.ActionGet, func(ctx context.Context, req interface{}) (string, error) {
				r := req.(*pb.GetPersonalContributionRequest)
				meID := apis.GetUserID(ctx)
				meOrgID := apis.GetOrgID(ctx)
				if meID != r.UserID {
					return "", fmt.Errorf("access denied: user mismatch")
				}
				if meOrgID != r.OrgID {
					return "", fmt.Errorf("access denied: org mismatch")
				}
				return req.(*pb.GetPersonalContributionRequest).OrgID, nil
			}),
		))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.dop.contribution.ContributionService" || ctx.Type() == pb.ContributionServiceServerType() || ctx.Type() == pb.ContributionServiceHandlerType():
		return p.contributionService
	}
	return p
}

func init() {
	servicehub.Register("erda.dop.contribution", &servicehub.Spec{
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
