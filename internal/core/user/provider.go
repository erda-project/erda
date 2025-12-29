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
	"fmt"
	"net/http"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/impl/iam"
	"github.com/erda-project/erda/internal/core/user/impl/uc"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	OAuthProvider string `default:"iam" file:"oauth_provider"`
}

type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register

	IAM         iam.Interface
	UC          uc.Interface
	userService common.Interface
}

func (p *provider) Init(_ servicehub.Context) error {
	switch p.Cfg.OAuthProvider {
	case domain.OAuthProviderIAM:
		p.userService = p.IAM
	case domain.OAuthProviderUC:
		p.userService = p.UC
	default:
		return fmt.Errorf("illegal oauth provider %s", p.Cfg.OAuthProvider)
	}

	p.Log.Infof("use oauth provider %s as user service", p.Cfg.OAuthProvider)
	if p.Register != nil {
		pb.RegisterUserServiceImp(p.Register, p.userService, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithDecoder(func(r *http.Request, out interface{}) error {
					switch body := out.(type) {
					// Rewrap payload: [{},{}] -> {"users": [{},{}]}
					case *pb.UserCreateRequest:
						var recv []*pb.UserCreateItem
						if err := encoding.DecodeRequest(r, &recv); err != nil {
							return err
						}
						body.Users = recv
						return nil
					default:
						return encoding.DecodeRequest(r, out)
					}
				}),
			),
		)
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
