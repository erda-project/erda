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

func Test_handleReverseRespHeader_ContentType(t *testing.T) {
	// content-type
	rw := mock.NewMockHTTPResponseWriter()
	resp := &http.Response{Header: make(http.Header)}

	// set rw header
	// set content-type
	rw.Header().Set(echo.HeaderContentType, "application/json")
	rw.Header().Set(echo.HeaderXRequestID, "1")

	// set resp header mock after proxy
	resp.Header.Set(echo.HeaderContentType, "application/json")
	resp.Header.Set(echo.HeaderXRequestID, "1")

	handleReverseRespHeader(rw, resp)

	assert.Equal(t, 2, len(rw.Header()))
	// will add echo.HeaderAccessControlAllowOrigin
	assert.Equal(t, 3, len(resp.Header))
}
