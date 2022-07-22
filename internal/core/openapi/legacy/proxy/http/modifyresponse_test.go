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

package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/pkg/mock"
)

func Test_getRealIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{
			name: "forwarded with space",
			ip:   "42.120.75.131,::ffff:10.112.1.1, 10.112.3.224",
			want: "42.120.75.131",
		},
		{
			name: "forwarded with no space",
			ip:   "42.120.75.131,::ffff:10.112.1.1,10.112.3.224",
			want: "42.120.75.131",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{Header: map[string][]string{echo.HeaderXForwardedFor: {tt.ip}}}
			if got := getRealIP(r); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
func Test_handleReverseRespHeader(t *testing.T) {
	// 1. simple header with one value
	rw := mock.NewMockHTTPResponseWriter()
	resp := &http.Response{Header: make(http.Header)}

	// set rw header
	// set cors
	rw.Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
	// set rid
	rw.Header().Set(echo.HeaderXRequestID, "1")

	// set resp header mock after proxy
	resp.Header.Set(echo.HeaderAccessControlAllowOrigin, "*")
	resp.Header.Set(echo.HeaderXRequestID, "1")

	handleReverseRespHeader(rw, resp)

	assert.Equal(t, 2, len(rw.Header()))
	assert.Equal(t, 0, len(resp.Header))

	// 2. complex header with multiple values
	// reset rw & resp
	rw = mock.NewMockHTTPResponseWriter()
	resp = &http.Response{Header: make(http.Header)}

	// set rw header
	// set cors
	rw.Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
	// set rid
	rw.Header().Set(echo.HeaderXRequestID, "1")
	// set multiple value header
	rw.Header().Set("mh", "1")
	rw.Header().Add("mh", "2")

	// set resp header mock after proxy
	resp.Header.Set(echo.HeaderAccessControlAllowOrigin, "*")
	resp.Header.Set(echo.HeaderXRequestID, "1")
	resp.Header.Set("mh", "1") // just "1" exists in rw

	handleReverseRespHeader(rw, resp)

	assert.Equal(t, 3, len(rw.Header()))
	assert.Equal(t, 0, len(resp.Header))

	// set resp header mock after proxy
	resp.Header = make(http.Header) // reset
	resp.Header.Set(echo.HeaderAccessControlAllowOrigin, "*")
	resp.Header.Set(echo.HeaderXRequestID, "1")
	resp.Header.Set("mh", "3") // just "3" not exists in rw

	handleReverseRespHeader(rw, resp)

	assert.Equal(t, 3, len(rw.Header()))
	assert.Equal(t, 1, len(resp.Header))
	assert.Equal(t, 1, len(resp.Header.Values("mh")))
	assert.Equal(t, "3", resp.Header.Get("mh"))

	// set resp header mock after proxy
	resp.Header = make(http.Header) // reset
	resp.Header.Set(echo.HeaderAccessControlAllowOrigin, "*")
	resp.Header.Set(echo.HeaderXRequestID, "1")
	resp.Header.Set("mh", "1") // "1" exists in rw
	resp.Header.Add("mh", "2") // "2" exists in rw
	resp.Header.Add("mh", "3") // and "3" not exists in rw
	resp.Header.Add("mh", "4") // and "4" not exists in rw

	handleReverseRespHeader(rw, resp)

	assert.Equal(t, 3, len(rw.Header()))
	assert.Equal(t, 1, len(resp.Header))
	assert.Equal(t, 2, len(resp.Header.Values("mh")))
	assert.Equal(t, "3", resp.Header.Get("mh"))
	assert.Equal(t, []string{"3", "4"}, resp.Header.Values("mh"))
}
