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

package credential

import (
	"net/http"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	akpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda-proto-go/msp/credential/pb"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

type provider struct {
	Cfg                  *config
	Register             transport.Register `autowired:"service-register"`
	credentialKeyService *accessKeyService
	AccessKeyService     akpb.AccessKeyServiceServer `autowired:erda.core.services.authentication.credentials.accesskey.accessKeyService"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.credentialKeyService = &accessKeyService{
		p: p,
	}
	if p.Register != nil {
		pb.RegisterAccessKeyServiceImp(p.Register, p.credentialKeyService, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithEncoder(func(rw http.ResponseWriter, r *http.Request, data interface{}) error {
					if resp, ok := data.(*apis.Response); ok && resp != nil {
						if data, ok := resp.Data.(*pb.DownloadAccessKeyFileResponse); ok {
							rw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
							rw.Header().Set("Pragma", "no-cache")
							rw.Header().Set("Expires", "0")
							rw.Header().Set("charset", "utf-8")
							rw.Header().Set("Content-Disposition", "attachment;filename=accessKey.csv")
							rw.Header().Set("Content-Type", "application/octet-stream")
							fluster := rw.(http.Flusher)
							rw.Write(data.Content)
							fluster.Flush()
							return nil
						}
					}
					return encoding.EncodeResponse(rw, r, data)
				})))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.credential.accessKeyService" || ctx.Type() == pb.AccessKeyServiceServerType() || ctx.Type() == pb.AccessKeyServiceHandlerType():
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
