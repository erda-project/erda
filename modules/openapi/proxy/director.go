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

package proxy

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/openapi/api"
	"github.com/erda-project/erda/modules/openapi/conf"
	"github.com/erda-project/erda/pkg/strutil"
)

const defautlFQDN string = "default.svc.cluster.local"

func NewDirector() func(*http.Request) {
	return func(r *http.Request) {
		path := r.URL.EscapedPath()
		path = strutil.Concat("/", strutil.TrimPrefixes(path, "/"))
		spec := api.API.Find(r)

		if spec == nil {
			// not found
			panic("should not be here")
		}
		r.URL.Scheme = spec.Scheme.String()
		if conf.UseK8S() {
			r.Host = spec.K8SHost
			r.URL.Host = spec.K8SHost
			erdaSystemFQDN := conf.ErdaSystemFQDN()
			if erdaSystemFQDN != "" && erdaSystemFQDN != defautlFQDN {
				host := replaceServiceName(erdaSystemFQDN, spec.K8SHost)
				r.Host = host
				r.URL.Host = host
			}
		} else {
			r.Host = spec.MarathonHost
			r.URL.Host = spec.MarathonHost
		}
		if r.Host == "" && spec.Custom == nil {
			logrus.Errorf("[alert][BUG] invalid host and spec.Custom, originHost=%v", spec.Host)
			r.Host = spec.Host
			r.URL.Host = spec.Host
		} else {
			r.Host = strutil.Concat(r.Host, ":", strconv.Itoa(spec.Port))
			r.URL.Host = strutil.Concat(r.URL.Host, ":", strconv.Itoa(spec.Port))
		}
		r.Header.Set("Origin-Path", path)
		r.URL.RawPath = spec.BackendPath.Fill(spec.Path.Vars(path))
		u, err := url.PathUnescape(r.URL.RawPath)
		if err != nil {
			logrus.Errorf("[alert] failed to unescape path: %v", r.URL.RawPath)
			r.URL.Path = r.URL.RawPath
		} else {
			r.URL.Path = u
		}
		// r.Header.Set("Origin", "http://"+r.Host)
	}
}

// genServiceName
func replaceServiceName(confFQDN, K8SHost string) string {
	svcName := strings.SplitN(K8SHost, ".", -1)[0]
	return svcName + "." + confFQDN
}
