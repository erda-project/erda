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

package plugins

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

func Test_mergeErdaIPConfig(t *testing.T) {
	type args struct {
		currentParaSignAuthConfig dto.MsePluginConfig
		updateParaSignAuthConfig  dto.MsePluginConfig
		updateForDisable          bool
	}

	cRules := make([]dto.Rules, 0)
	uRules := make([]dto.Rules, 0)

	cRules = append(cRules, dto.Rules{
		MatchRoute: []string{MseDefaultRouteName},
		IPSource:   common.MseErdaIpSourceXRealIP,
		IpAclType:  common.MseErdaIpAclWhite,
		IpAclList:  []string{"1.1.1.1", "1.2.3.0/24"},
	})
	cRules = append(cRules, dto.Rules{
		MatchRoute: []string{"test-route"},
		IPSource:   common.MseErdaIpSourceXRealIP,
		IpAclType:  common.MseErdaIpAclWhite,
		IpAclList:  []string{"1.1.1.1", "1.2.3.0/24"},
	})

	uRules = append(uRules, dto.Rules{
		MatchRoute: []string{"test-route"},
		IPSource:   common.MseErdaIpSourceXRealIP,
		IpAclType:  common.MseErdaIpAclWhite,
		IpAclList:  []string{"1.1.1.1", "1.2.3.0/24"},
	})

	tests := []struct {
		name    string
		args    args
		want    dto.MsePluginConfig
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				currentParaSignAuthConfig: dto.MsePluginConfig{
					Rules: cRules,
				},
				updateParaSignAuthConfig: dto.MsePluginConfig{
					Rules: uRules,
				},
				updateForDisable: true,
			},
			want: dto.MsePluginConfig{
				Rules: []dto.Rules{{
					MatchRoute: []string{MseDefaultRouteName},
					IPSource:   common.MseErdaIpSourceXRealIP,
					IpAclType:  common.MseErdaIpAclWhite,
					IpAclList:  []string{"1.1.1.1", "1.2.3.0/24"},
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeErdaIPConfig(tt.args.currentParaSignAuthConfig, tt.args.updateParaSignAuthConfig, tt.args.updateForDisable)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeErdaIPConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeErdaIPConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getErdaIPSourceConfig(t *testing.T) {
	type args struct {
		config map[string]interface{}
	}

	tests := []struct {
		name          string
		args          args
		wantIpSource  string
		wantIpAclType string
		wantIpAclList []string
		wantDisable   bool
		wantErr       bool
	}{
		{
			name: "Test_01",
			args: args{config: map[string]interface{}{
				common.MseErdaIpIpSource:    common.MseErdaIpSourceXRealIP,
				common.MseErdaIpAclType:     common.MseErdaIpAclBlack,
				common.MseErdaIpAclList:     []string{"10.10.10.10", "11.12.13.0/24"},
				common.MseErdaIpRouteSwitch: false,
			}},
			wantIpSource:  common.MseErdaIpSourceXRealIP,
			wantIpAclType: common.MseErdaIpAclBlack,
			wantIpAclList: []string{"10.10.10.10", "11.12.13.0/24"},
			wantDisable:   true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIpSource, gotIpAclType, gotIpAclList, gotDisable, err := getErdaIPSourceConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("getErdaIPSourceConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIpSource != tt.wantIpSource {
				t.Errorf("getErdaIPSourceConfig() gotIpSource = %v, want %v", gotIpSource, tt.wantIpSource)
			}
			if gotIpAclType != tt.wantIpAclType {
				t.Errorf("getErdaIPSourceConfig() gotIpAclType = %v, want %v", gotIpAclType, tt.wantIpAclType)
			}

			if gotDisable != tt.wantDisable {
				t.Errorf("getErdaIPSourceConfig() gotDisable = %v, want %v", gotDisable, tt.wantDisable)
			}

			if !reflect.DeepEqual(gotIpAclList, tt.wantIpAclList) {
				t.Errorf("getErdaIPSourceConfig() gotIpAclList = %v, want %v", gotIpAclList, tt.wantIpAclList)
			}
		})
	}
}
