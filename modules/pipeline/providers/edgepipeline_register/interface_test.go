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

package edgepipeline_register

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/bmizerany/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func Test_parseDialerEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
		wantErr  bool
	}{
		{
			name:     "invalid endpoint",
			endpoint: "xxx",
			want:     "xxx",
			wantErr:  false,
		},
		{
			name:     "http endpoint",
			endpoint: "http://cluster-dialer:80",
			want:     "ws://cluster-dialer:80",
			wantErr:  false,
		},
		{
			name:     "https endpoint",
			endpoint: "https://cluster-dialer:80",
			want:     "wss://cluster-dialer:80",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		p := &provider{
			Cfg: &Config{
				IsEdge:              true,
				ClusterDialEndpoint: tt.endpoint,
			},
		}
		got, err := p.parseDialerEndpoint()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. provider.parseDialerEndpoint() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. provider.parseDialerEndpoint() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSourceWhiteList(t *testing.T) {
	p := &provider{
		Cfg: &Config{
			AllowedSources: []string{"cdp-", "recommend-"},
		},
	}
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "cdp source",
			src:  "cdp-123",
			want: true,
		},
		{
			name: "default source",
			src:  "default",
			want: false,
		},
		{
			name: "dice source",
			src:  "dice",
			want: false,
		},
		{
			name: "valid source with prefix",
			src:  "recommend-123",
			want: true,
		},
		{
			name: "invalid source with prefix",
			src:  "invalid-123",
			want: false,
		},
	}
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(p.bdl), "IsClusterDialerClientRegistered", func(_ *bundle.Bundle, _ string, _ string) (bool, error) {
		return true, nil
	})
	defer patch.Unpatch()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.ShouldDispatchToEdge(tt.src, "dev"); got != tt.want {
				t.Errorf("sourceWhiteList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_GetEdgeBundleByClusterName(t *testing.T) {
	type fields struct {
	}
	type args struct {
		clusterName string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantOpenApiAddr string
		wantErr         bool
	}{
		{
			name:   "test pipeline addr",
			fields: fields{},
			args: args{
				clusterName: "test",
			},
			wantOpenApiAddr: "test",
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}

			var detail = apistructs.ClusterDialerClientDetail{
				apistructs.ClusterDialerDataKeyPipelineAddr: tt.wantOpenApiAddr,
			}
			patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(detail), "Get", func(detail apistructs.ClusterDialerClientDetail, key apistructs.ClusterDialerClientDetailKey) string {
				assert.Equal(t, detail[key], tt.wantOpenApiAddr)
				return detail[key]
			})
			defer patch1.Unpatch()

			var bdl = &bundle.Bundle{}
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetClusterDialerClientData", func(bdl *bundle.Bundle, clientType string, clusterKey string) (apistructs.ClusterDialerClientDetail, error) {
				return detail, nil
			})
			p.bdl = bdl
			defer patch.Unpatch()

			_, err := p.GetEdgeBundleByClusterName(tt.args.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEdgeBundleByClusterName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
