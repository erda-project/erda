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

package proxy

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
)

func TestPolicy_CreateDefaultConfig(t *testing.T) {
	type fields struct {
		BasePolicy apipolicy.BasePolicy
	}
	type args struct {
		gatewayProvider string
		ctx             map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   apipolicy.PolicyDto
	}{
		{
			name: "Test_01",
			fields: fields{
				BasePolicy: apipolicy.BasePolicy{
					PolicyName: apipolicy.Policy_Engine_Proxy,
				},
			},
			args: args{
				gatewayProvider: "",
				ctx:             nil,
			},
			want: &PolicyDto{
				ReqBuffer:         true,
				RespBuffer:        true,
				ClientReqLimit:    100,
				ClientReqTimeout:  60,
				ClientRespTimeout: 60,
				ProxyReqTimeout:   60,
				ProxyRespTimeout:  60,
				HostPassthrough:   false,
				SSLRedirect:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := Policy{
				BasePolicy: tt.fields.BasePolicy,
			}
			if got := policy.CreateDefaultConfig(tt.args.gatewayProvider, tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateDefaultConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
