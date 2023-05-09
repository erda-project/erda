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

package ip

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
)

func TestPolicy_buildMSEPluginReq(t *testing.T) {
	type fields struct {
		BasePolicy apipolicy.BasePolicy
	}
	type args struct {
		dto      *PolicyDto
		zoneName string
	}
	tSwitch := true
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *providerDto.PluginReqDto
	}{
		{
			name: "Test_01",
			fields: fields{
				BasePolicy: apipolicy.BasePolicy{
					PolicyName: apipolicy.Policy_Engine_IP,
				},
			},
			args: args{
				dto: &PolicyDto{
					BaseDto: apipolicy.BaseDto{
						Switch: tSwitch,
						Global: false,
					},
					IpSource:  X_REAL_IP,
					IpAclType: ACL_WHITE,
					IpAclList: []string{"1.1.1.1", "1.2.3.0/24"},
				},
				zoneName: "zone",
			},
			want: &providerDto.PluginReqDto{
				Name: mseCommon.MsePluginIP,
				Config: map[string]interface{}{
					mseCommon.MseErdaIpIpSource:    mseCommon.MseErdaIpSourceXRealIP,
					mseCommon.MseErdaIpAclType:     string(ACL_WHITE),
					mseCommon.MseErdaIpAclList:     []string{"1.1.1.1", "1.2.3.0/24"},
					mseCommon.MseErdaIpRouteSwitch: tSwitch,
				},
				Enabled:  &tSwitch,
				ZoneName: "zone",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := Policy{
				BasePolicy: tt.fields.BasePolicy,
			}
			if got := policy.buildMSEPluginReq(tt.args.dto, tt.args.zoneName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildPluginReq() = %v, want %v", got, tt.want)
			}
		})
	}
}
