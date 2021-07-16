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

package metric

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	api "github.com/erda-project/erda/pkg/common/httpapi"
)

func (p *provider) proxy(hostpath string, header http.Header, params url.Values, rw http.ResponseWriter, r *http.Request) error {
	target, err := url.Parse(hostpath)
	if err != nil {
		return err
	}
	path, rawpath := target.Path, target.EscapedPath()
	rp := &httputil.ReverseProxy{Director: func(req *http.Request) {
		for k, vals := range header {
			req.Header.Del(k)
			for _, val := range vals {
				req.Header.Add(k, val)
			}
		}
		query := req.URL.Query()
		for k, vals := range params {
			query.Del(k)
			for _, val := range vals {
				query.Add(k, val)
			}
		}
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = path, rawpath
		req.URL.RawQuery = query.Encode()
		p.Log.Debugf("proxy %s %s -> %s", req.Method, rawpath, req.URL)
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}}
	rp.ServeHTTP(rw, r)
	return nil
}

func (p *provider) proxyMonitor(path string, params url.Values, rw http.ResponseWriter, r *http.Request) interface{} {
	err := p.proxy(fmt.Sprintf("%s%s", p.Cfg.MonitorURL, path), nil, params, rw, r)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return nil
}
