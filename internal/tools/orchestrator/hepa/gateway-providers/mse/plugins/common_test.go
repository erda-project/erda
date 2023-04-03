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

	mseclient "github.com/alibabacloud-go/mse-20190531/v3/client"

	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
)

func TestUpdatePluginConfigWhenDeleteConsumer(t *testing.T) {
	type args struct {
		pluginName   string
		consumerName string
		config       interface{}
	}
	//config.([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList)
	cfList := make([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, 0)
	config := "consumers: \n- credential: 5d1a401b3aef4ee5bc39a99654672f91\n  name: aaa\nkeys:\n- apikey\n- x-api-key\n_rules_:\n- _match_route_:\n  - project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3\n  allow:\n  - aaa\n  \n"
	var configLevel int32 = 0
	cfList = append(cfList, &mseclient.GetPluginConfigResponseBodyDataGatewayConfigList{
		Config:      &config,
		ConfigLevel: &configLevel,
	})

	wantCFList := make([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, 0)
	wantConfig := "consumers:\n    - name: aaa\n      credential: 5d1a401b3aef4ee5bc39a99654672f91\nkeys:\n    - apikey\n    - x-api-key\n_rules_:\n    - _match_route_:\n        - project-5846-test-dice-test-5846-api-754bbce0af034774ac2b8f74c7e070a6-0149aa-058a61f1-ccaae83a5fe7846d18d4ac80940a8fdc3\n      allow:\n        - aaa\n"
	wantCFList = append(wantCFList, &mseclient.GetPluginConfigResponseBodyDataGatewayConfigList{
		Config:      &wantConfig,
		ConfigLevel: &configLevel,
	})
	tests := []struct {
		name    string
		args    args
		want    []*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				pluginName:   mseCommon.MsePluginKeyAuth,
				consumerName: "xyz",
				config:       cfList,
			},
			want:    wantCFList,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdatePluginConfigWhenDeleteConsumer(tt.args.pluginName, tt.args.consumerName, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePluginConfigWhenDeleteConsumer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdatePluginConfigWhenDeleteConsumer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
