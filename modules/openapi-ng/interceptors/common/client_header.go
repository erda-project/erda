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

package common

import (
	"net/http"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// +provider
type provider struct {
	Cfg *interceptors.Config
}

func (p *provider) List() []*interceptors.Interceptor {
	return []*interceptors.Interceptor{{Order: p.Cfg.Order, Wrapper: p.Interceptor}}
}

func (p *provider) Interceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		r.Header.Del(httputil.InternalHeader)
		r.Header.Del(httputil.ClientIDHeader)
		r.Header.Del(httputil.ClientNameHeader)
		h(rw, r)
	}
}

func init() {
	servicehub.Register("openapi-interceptor-filter-client-header", &servicehub.Spec{
		Services:   []string{"openapi-interceptor-filter-client-header"},
		ConfigFunc: func() interface{} { return &interceptors.Config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
