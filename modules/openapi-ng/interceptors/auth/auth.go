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
