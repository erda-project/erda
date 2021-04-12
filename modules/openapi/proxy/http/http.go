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

package http

import (
	"net/http"
	"net/http/httputil"
	"runtime/debug"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/openapi/api"
)

func NewReverseProxy(director func(*http.Request),
	modifyResponse func(*http.Response) error) http.Handler {
	return &httputil.ReverseProxy{
		FlushInterval:  -1,
		Director:       director,
		ModifyResponse: modifyResponse,
	}
}

type ReverseProxyWithCustom struct {
	reverseProxy http.Handler
}

func NewReverseProxyWithCustom(director func(*http.Request),
	modifyResponse func(*http.Response) error) http.Handler {
	r := NewReverseProxy(director, modifyResponse)
	return &ReverseProxyWithCustom{reverseProxy: r}
}

func (r *ReverseProxyWithCustom) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logrus.Errorf("[alert] openapi http proxy recover from panic: %v", err)
		}
	}()
	spec := api.API.Find(req)
	if spec != nil && spec.Custom != nil {
		spec.Custom(rw, req)
		return
	}
	r.reverseProxy.ServeHTTP(rw, req)
}
