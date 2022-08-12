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

package httpserver

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	infrahttpserver "github.com/erda-project/erda-infra/providers/httpserver"
)

func Test_getVars(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name          string
		args          args
		beforeGetVars func(r *http.Request)
		wantVars      map[string]string
	}{
		{
			name: "no vars",
			args: args{
				r: func() *http.Request {
					return &http.Request{}
				}(),
			},
			beforeGetVars: nil,
			wantVars:      nil,
		},
		{
			name: "vars from legacyhttpserver",
			args: args{
				r: func() *http.Request {
					r := &http.Request{}
					r = mux.SetURLVars(r, map[string]string{"id": "1"})
					return r
				}(),
			},
			beforeGetVars: nil,
			wantVars:      map[string]string{"id": "1"},
		},
		{
			name: "vars from infrahttpserver",
			args: args{
				r: func() *http.Request {
					r := &http.Request{}
					return r
				}(),
			},
			beforeGetVars: func(r *http.Request) {
				mockedInfrahttpserverVars := func(r *http.Request) map[string]string {
					return map[string]string{"id": "1"}
				}
				infrahttpserverVars = mockedInfrahttpserverVars
			},
			wantVars: map[string]string{"id": "1"},
		},
		{
			name: "vars encoded",
			args: args{
				r: func() *http.Request {
					r := &http.Request{}
					r = mux.SetURLVars(r, map[string]string{"id": url.QueryEscape("Asia/Shanghai")})
					return r
				}(),
			},
			beforeGetVars: nil,
			wantVars:      map[string]string{"id": "Asia/Shanghai"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset infrahttpserverVars for each case
			infrahttpserverVars = infrahttpserver.Vars
			if tt.beforeGetVars != nil {
				tt.beforeGetVars(tt.args.r)
			}
			assert.Equalf(t, tt.wantVars, getVars(tt.args.r), "getVars(%v)", tt.args.r)
		})
	}
}
