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
	"os"
	"regexp"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	proxyURLs       = map[string]*url.URL{}
	proxyREs        = map[string]*regexp.Regexp{}
	proxyHttpServer = goproxy.NewProxyHttpServer()
)

func init() {
	for _, s := range strings.Split(os.Getenv("PROXY_SERVICES"), ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		k := strings.Replace(strings.ToUpper(s), "-", "_", -1)
		if v := os.Getenv(k + "_BACKEND_URL"); v != "" {
			u, err := url.Parse(v)
			if err != nil || !(u.Scheme == "http" || u.Scheme == "https") || u.Host == "" {
				logrus.Fatalln("backend url invalid:", s)
			}
			logrus.Infoln(s, u.String())
			proxyURLs[s] = u
			continue
		}
		if v := os.Getenv(k + "_BACKEND_RE"); v != "" {
			r, err := regexp.Compile(v)
			if err != nil {
				logrus.Fatalln("backend re invalid:", s)
			}
			logrus.Infoln(s, r.String())
			proxyREs[s] = r
			continue
		}
		logrus.Fatalln("no backend:", s)
	}
}

func ProxyService(w http.ResponseWriter, r *http.Request) {
	s := mux.Vars(r)["service"]
	u, ok := proxyURLs[s]
	if !ok {
		re, ok := proxyREs[s]
		if !ok {
			http.Error(w, "no "+s, http.StatusBadGateway)
			return
		}
		u, _ = url.Parse(r.Header.Get(strings.Replace(s, "_", "-", -1) + "-BACKEND-URL"))
		if u == nil || !(u.Scheme == "http" || u.Scheme == "https") || !re.MatchString(u.Host) {
			http.Error(w, "illegal "+s, http.StatusBadRequest)
			return
		}
	}
	p := strings.TrimPrefix(r.URL.Path, "/api/proxy/"+s)
	if p == "" || p == "/" {
		p = u.Path
	} else if !strings.HasPrefix(p, "/") {
		http.Error(w, "no "+s+" path", http.StatusBadGateway)
		return
	} else {
		p = strings.TrimSuffix(u.Path, "/") + p
	}
	r2 := new(http.Request)
	*r2 = *r
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	r2.URL.Scheme = u.Scheme
	r2.URL.Host = u.Host
	r2.Host = u.Host
	r2.URL.Path = p
	// RawPath? RequestURI?
	logrus.Infoln("proxy", s, r.RemoteAddr, r2.URL.String())
	proxyHttpServer.ServeHTTP(w, r2)
}

// deprecated
func ProxyFPS(w http.ResponseWriter, r *http.Request) {
	const s = "fps"
	u, ok := proxyURLs[s]
	if !ok {
		http.Error(w, "no fps", http.StatusBadGateway)
		return
	}
	r2 := new(http.Request)
	*r2 = *r
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	r2.URL.Scheme = u.Scheme
	r2.URL.Host = u.Host
	r2.Host = u.Host
	logrus.Infoln("proxy", s, r.RemoteAddr, r2.URL.String())
	proxyHttpServer.ServeHTTP(w, r2)
}
