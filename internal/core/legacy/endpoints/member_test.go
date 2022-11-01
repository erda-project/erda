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

package endpoints

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/erda-project/erda/pkg/http/httputil"
)

func Test_getMemberQueryParamForDesensitize(t *testing.T) {
	tests := []struct {
		name                  string
		req                   *http.Request
		wantDesensitizeEmail  bool
		wantDesensitizeMobile bool
	}{
		{
			name:                  "not specified",
			req:                   &http.Request{URL: &url.URL{}},
			wantDesensitizeEmail:  true,
			wantDesensitizeMobile: true,
		},
		{
			name:                  "email=true, mobile=true",
			req:                   &http.Request{URL: &url.URL{RawQuery: "desensitizeEmail=true&desensitizeMobile=true"}},
			wantDesensitizeEmail:  true,
			wantDesensitizeMobile: true,
		},
		{
			name: "email=true, mobile=false",
			req: &http.Request{
				URL:    &url.URL{RawQuery: "desensitizeEmail=true&desensitizeMobile=false"},
				Header: http.Header{httputil.InternalHeader: []string{"ut"}},
			},
			wantDesensitizeEmail:  true,
			wantDesensitizeMobile: false,
		},
		{
			name: "email=false, mobile=true",
			req: &http.Request{
				URL:    &url.URL{RawQuery: "desensitizeEmail=false&desensitizeMobile=true"},
				Header: http.Header{httputil.InternalHeader: []string{"ut"}},
			},
			wantDesensitizeEmail:  false,
			wantDesensitizeMobile: true,
		},
		{
			name: "email=false, mobile=false",
			req: &http.Request{
				URL:    &url.URL{RawQuery: "desensitizeEmail=false&desensitizeMobile=false"},
				Header: http.Header{httputil.InternalHeader: []string{"ut"}},
			},
			wantDesensitizeEmail:  false,
			wantDesensitizeMobile: false,
		},
		{
			name: "not internal invoke, email=false, mobile=false",
			req: &http.Request{
				URL:    &url.URL{RawQuery: "desensitizeEmail=false&desensitizeMobile=false"},
				Header: http.Header{httputil.InternalHeader: nil},
			},
			wantDesensitizeEmail:  true,
			wantDesensitizeMobile: true,
		},
		{
			name: "invalid flag value, use default true",
			req: &http.Request{
				URL:    &url.URL{RawQuery: "desensitizeEmail=what&desensitizeMobile=wrong"},
				Header: http.Header{httputil.InternalHeader: []string{"ut"}},
			},
			wantDesensitizeEmail:  true,
			wantDesensitizeMobile: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDesensitizeEmail, gotDesensitizeMobile := getMemberQueryParamForDesensitize(tt.req)
			if gotDesensitizeEmail != tt.wantDesensitizeEmail {
				t.Errorf("getMemberQueryParamForDesensitize() gotDesensitizeEmail = %v, want %v", gotDesensitizeEmail, tt.wantDesensitizeEmail)
			}
			if gotDesensitizeMobile != tt.wantDesensitizeMobile {
				t.Errorf("getMemberQueryParamForDesensitize() gotDesensitizeMobile = %v, want %v", gotDesensitizeMobile, tt.wantDesensitizeMobile)
			}
		})
	}
}
