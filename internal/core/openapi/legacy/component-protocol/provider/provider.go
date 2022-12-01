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

package provider

import (
	"net/http"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-proto-go/core/openapi/component-protocol/pb"
	"github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol/proxy"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

type provider struct {
	Cfg      *config
	Register transport.Register

	s *service
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.s = &service{p: p}
	if p.Register != nil {
		pb.RegisterOpenapiComponentProtocolImp(p.Register, p.s, apis.Options(),
			transport.WithHTTPOptions(
				// overwrite default decoder
				transhttp.WithDecoder(func(req *http.Request, data interface{}) error {
					return nil
				}),
				// actual logic is in custom encoder
				transhttp.WithEncoder(func(rw http.ResponseWriter, req *http.Request, _ interface{}) error {
					proxy.ProxyOrLegacy(rw, req)
					return nil
				}),
			),
		)
	}
	return nil
}

func init() {
	servicehub.Register("openapi-component-protocol", &servicehub.Spec{
		Services:   []string{"openapi-component-protocol"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
