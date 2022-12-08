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

package response

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors"
)

type config struct {
	Order         int
	XFrameOptions string `file:"x_frame_options" default:"DENY" env:"X_FRAME_OPTIONS" desc:"x-frame-options, optional values: DENY, SAMEORIGIN, ALLOW-FROM"`
}

// +provider
type provider struct {
	Cfg *config
}

var _ interceptors.Interface = (*provider)(nil)

func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *provider) List() []*interceptors.Interceptor {
	return []*interceptors.Interceptor{{Order: p.Cfg.Order, Wrapper: p.Interceptor}}
}

func (p *provider) Interceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
		rw.Header().Set(echo.HeaderXFrameOptions, p.Cfg.XFrameOptions)
		h(rw, r)
	}
}

func init() {
	servicehub.Register("openapi-interceptor-set-response-header", &servicehub.Spec{
		Services:   []string{"openapi-interceptor-set-response-header"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
