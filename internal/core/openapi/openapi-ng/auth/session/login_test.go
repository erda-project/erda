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

package oauth

import (
	"testing"
)

func TestGetUCRedirectHost(t *testing.T) {
	tests := []struct {
		name       string
		referer    string
		host       string
		config     config
		wantResult string
	}{
		{
			name:    "Matching referer with UCRedirectAddrs",
			referer: "https://erda.cloud/login",
			host:    "erda.cloud",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud"},
			},
			wantResult: "openapi.erda.cloud",
		},
		{
			name:    "Non-matching referer with UCRedirectAddrs",
			referer: "https://erda.cloud/login",
			host:    "erda.cloud",
			config: config{
				UCRedirectAddrs: []string{"fake.erda.cloud"},
			},
			wantResult: "fake.erda.cloud",
		},
		{
			name:    "Host with port number",
			referer: "https://erda.cloud:8080/login",
			host:    "erda.cloud:8080",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud:8080"},
			},
			wantResult: "openapi.erda.cloud:8080",
		},
		{
			name:    "Empty host",
			referer: "https://erda.cloud:8080/login",
			host:    "",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud:8080"},
			},
			wantResult: "openapi.erda.cloud:8080",
		},
		{
			name:    "Referer and host have different domains",
			referer: "https://erda.cloud/login",
			host:    "another.com",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud"},
			},
			wantResult: "openapi.erda.cloud",
		},
		{
			name:    "Empty host and Referer with diff port",
			referer: "https://erda.cloud:8080/login",
			host:    "",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud:9090"},
			},
			wantResult: "openapi.erda.cloud:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				Cfg: &tt.config,
			}
			got := p.getUCRedirectHost(tt.referer, tt.host)
			if got != tt.wantResult {
				t.Errorf("getUCRedirectHost() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
