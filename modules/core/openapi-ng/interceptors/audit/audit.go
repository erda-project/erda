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

package audit

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/core/openapi-ng/interceptors"
)

type config struct {
	Order int `default:"1000"`
}

type provider struct {
	Cfg *config
}

var _ interceptors.Interface = (*provider)(nil)

func (p *provider) List() []*interceptors.Interceptor {
	return []*interceptors.Interceptor{{Order: p.Cfg.Order, Wrapper: p.Interceptor}}
}

func getHeaderValue(v string) string {
	i := strings.Index(v, ";")
	if i < 0 {
		return strings.TrimSpace(strings.ToLower(v))
	}
	return strings.TrimSpace(strings.ToLower(v[0:i]))
}

type wrapedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	err        error
}

func (rw *wrapedResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *wrapedResponseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	if err != nil {
		rw.err = err
	}
	return n, err
}

func (rw *wrapedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("%T does not implement http.Hijacker", rw.ResponseWriter)
}

func (rw *wrapedResponseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (p *provider) Interceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		resp := &wrapedResponseWriter{
			ResponseWriter: rw,
		}
		h(resp, r)
		header := resp.Header()
		if resp.statusCode < 300 {
			if audit, _ := strconv.ParseBool(header.Get("Audit")); audit {
				for key, vals := range header {
					key = strings.ToLower(key)
					if strings.HasPrefix(key, "audit-") {
						// TODO: do audit
						for _, val := range vals {
							_ = val
						}
					}
				}
			}
		}
	}
}

func init() {
	servicehub.Register("openapi-interceptor-audit", &servicehub.Spec{
		Services:   []string{"openapi-interceptor-audit"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
