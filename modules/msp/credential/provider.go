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

package credential

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	akpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda-proto-go/msp/credential/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	"net/http"
)

type config struct {
}

type provider struct {
	Cfg                  *config
	Register             transport.Register `autowired:"service-register"`
	credentialKeyService *accessKeyService
	AccessKeyService     akpb.AccessKeyServiceServer `autowired:erda.core.services.authentication.credentials.accesskey.AccessKeyService"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.credentialKeyService = &accessKeyService{
		p: p,
	}
	if p.Register != nil {
		pb.RegisterAccessKeyServiceImp(p.Register, p.credentialKeyService, apis.Options(), transport.WithHTTPOptions(
			transhttp.WithEncoder(func(rw http.ResponseWriter, r *http.Request, data interface{}) error {
				if resp, ok := data.(*pb.DownloadAccessKeyFileResponse); ok {
					rw.Write(resp.Data)
				}
				return encoding.EncodeResponse(rw, r, data)
			})))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.credential.AccessKeyService" || ctx.Type() == pb.AccessKeyServiceServerType() || ctx.Type() == pb.AccessKeyServiceHandlerType():
		return p.AccessKeyService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.credential", &servicehub.Spec{
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
