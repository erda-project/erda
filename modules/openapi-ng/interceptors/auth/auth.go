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

package auth

import (
	"fmt"
	"net/http"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type config struct {
	Order int `default:"10"`
}

// +provider
type provider struct {
	Cfg *config
}

func (p *provider) List() []*interceptors.Interceptor {
	return []*interceptors.Interceptor{{Order: p.Cfg.Order, Wrapper: p.Interceptor}}
}

func (p *provider) Interceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// TODO .
		r.Header.Del(httputil.UserHeader)
		r.Header.Del(httputil.OrgHeader)
		fmt.Println("TODO auth ...")
		h(rw, r)
	}
}

func init() {
	servicehub.Register("openapi-interceptor-auth", &servicehub.Spec{
		Services:   []string{"openapi-interceptor-auth"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
