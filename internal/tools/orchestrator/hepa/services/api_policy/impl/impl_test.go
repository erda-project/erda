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

package impl

import (
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
)

func Test_hasDuplicatedConfig(t *testing.T) {
	type args struct {
		p1 apipolicy.PolicyConfig
		p2 apipolicy.PolicyConfig
	}

	proxy_next_upstream := "error timeout http_429 non_idempotent"
	annos1 := make(map[string]*string)
	annos1["proxy-next-upstream"] = &proxy_next_upstream
	proxy_read_timeout := "120"
	annos1["proxy-read-timeout"] = &proxy_read_timeout

	//ls1 := "\n  proxy_intercept_errors on;\n"

	annos2 := make(map[string]*string)
	annos2["proxy-next-upstream"] = &proxy_next_upstream
	ls1 := "\n  send_timeout 120s; \n"
	ls2 := "\n  proxy_intercept_errors on;\n"

	pn1 := &apipolicy.IngressAnnotation{
		Annotation:      annos1,
		LocationSnippet: &ls1,
	}

	pn2 := &apipolicy.IngressAnnotation{
		Annotation:      annos2,
		LocationSnippet: &ls2,
	}

	ls3 := "\n  proxy_read_timeout 600s;\n   proxy_connect_timeout 600s;\n   proxy_send_timeout 600s;\n"
	pn3 := &apipolicy.IngressAnnotation{
		Annotation:      nil,
		LocationSnippet: &ls3,
	}

	co := make(map[string]*string)
	co["proxy-next-upstream"] = &proxy_next_upstream
	pc2 := &apipolicy.IngressController{
		ConfigOption: co,
	}

	ms := "\n     proxy_next_upstream error timeout http_581 non_idempotent;\n" +
		"     more_set_headers  'Access-Control-Allow-Methods: GET, PUT, POST, DELETE, PATCH, OPTIONS';\n" +
		"     more_set_headers  'Access-Control-Allow-Headers: $http_access_control_request_headers'; \n"
	pc3 := &apipolicy.IngressController{
		MainSnippet: &ms,
	}

	hs := `
location @LIMIT-xxxxxx {
    log_by_lua_block {
        plugins.run()
    }
    more_set_headers 'Access-Control-Allow-Origin: $from_request_origin_or_referer';
    more_set_headers 'Access-Control-Allow-Methods: GET, PUT, POST, DELETE, PATCH, OPTIONS';
    more_set_headers 'Access-Control-Allow-Headers: $http_access_control_request_headers';
    more_set_headers 'Access-Control-Allow-Credentials: true';
    more_set_headers 'Access-Control-Max-Age: 86400';
    more_set_headers 'Content-Type: text/plain charset=UTF-8';
    return 200;
}
`
	pc4 := &apipolicy.IngressController{
		HttpSnippet: &hs,
	}

	pc5 := &apipolicy.IngressController{
		ServerSnippet: &hs,
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				p1: apipolicy.PolicyConfig{
					IngressAnnotation: pn1,
					IngressController: nil,
				},
				p2: apipolicy.PolicyConfig{
					IngressAnnotation: pn2,
					IngressController: nil,
				},
			},
			want:    "proxy_next_upstream=error timeout http_429 non_idempotent",
			wantErr: false,
		},
		{
			name: "Test_02",
			args: args{
				p1: apipolicy.PolicyConfig{
					IngressAnnotation: pn1,
					IngressController: nil,
				},
				p2: apipolicy.PolicyConfig{
					IngressAnnotation: nil,
					IngressController: pc2,
				},
			},
			want:    "proxy_next_upstream=error timeout http_429 non_idempotent",
			wantErr: false,
		},
		{
			name: "Test_03",
			args: args{
				p1: apipolicy.PolicyConfig{
					IngressAnnotation: pn1,
					IngressController: nil,
				},
				p2: apipolicy.PolicyConfig{
					IngressAnnotation: nil,
					IngressController: pc3,
				},
			},
			want:    "proxy_next_upstream=error timeout http_581 non_idempotent",
			wantErr: false,
		},
		{
			name: "Test_04",
			args: args{
				p1: apipolicy.PolicyConfig{
					IngressAnnotation: pn1,
					IngressController: nil,
				},
				p2: apipolicy.PolicyConfig{
					IngressAnnotation: nil,
					IngressController: pc4,
				},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Test_05",
			args: args{
				p1: apipolicy.PolicyConfig{
					IngressAnnotation: pn1,
					IngressController: nil,
				},
				p2: apipolicy.PolicyConfig{
					IngressAnnotation: nil,
					IngressController: pc5,
				},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Test_06",
			args: args{
				p1: apipolicy.PolicyConfig{
					IngressAnnotation: pn1,
					IngressController: nil,
				},
				p2: apipolicy.PolicyConfig{
					IngressAnnotation: pn3,
				},
			},
			want:    "proxy_read_timeout=600s",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hasDuplicatedConfig(tt.args.p1, tt.args.p2)
			if (err != nil) != tt.wantErr {
				t.Errorf("hasDuplicatedConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("hasDuplicatedConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateCustomNginxConf(t *testing.T) {
	type args struct {
		category string
		config   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test_01",
			args: args{
				category: "custom",
				config:   "/n    xxxx 123;\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateCustomNginxConf(tt.args.category, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("validateCustomNginxConf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
