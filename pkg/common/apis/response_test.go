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

package apis

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

type mockResponseWriter struct {
	data []byte
}

func (m *mockResponseWriter) Header() http.Header {
	return map[string][]string{}
}

func (m *mockResponseWriter) Write(bytes []byte) (int, error) {
	m.data = bytes
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {

}

type transhttpError struct{}

func (t *transhttpError) Error() string {
	return "transport error"
}

func (t *transhttpError) HTTPStatus() int {
	return http.StatusBadRequest
}

func Test_encodeError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "api error",
			args: args{
				err: errorresp.New(
					errorresp.WithMessage("api error msg"),
					errorresp.WithTemplateMessage("err1", "val1"),
					errorresp.WithCode(400, "400"),
				),
			},
			want: []byte(`{"success":false,"err":{"code":"400","msg":"api error msg: val1","ctx":null}}
`),
		},
		{
			name: "transhttp error",
			args: args{
				err: &transhttpError{},
			},
			want: []byte(`{"success":false,"err":{"code":"400","msg":"transport error","ctx":"/api"}}
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, r := &mockResponseWriter{}, &http.Request{URL: &url.URL{Path: "/api"}}
			encodeError(w, r, tt.args.err)
			if string(w.data) != string(tt.want) {
				t.Errorf("encodeError() = %v, want %v", w.data, tt.want)
			}
		})
	}
}
