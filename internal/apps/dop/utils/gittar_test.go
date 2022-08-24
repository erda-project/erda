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

package utils

import (
	"testing"
)

func Test_getCenterOrEdgeURL(t *testing.T) {
	type args struct {
		diceCluster    string
		requestCluster string
		center         string
		saas           string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with saas",
			args: args{
				diceCluster:    "erda",
				requestCluster: "erda1",
				center:         "http://gittar.default.svc.cluster.local",
				saas:           "https://gittar.erda.cloud",
			},
			want: "https://gittar.erda.cloud",
		},
		{
			name: "test with center1",
			args: args{
				diceCluster:    "erda",
				requestCluster: "erda",
				center:         "http://gittar.default.svc.cluster.local",
				saas:           "",
			},
			want: "http://gittar.default.svc.cluster.local",
		},
		{
			name: "test with cente2",
			args: args{
				diceCluster:    "erda",
				requestCluster: "erda",
				center:         "http://gittar.default.svc.cluster.local",
				saas:           "https://gittar.erda.cloud",
			},
			want: "http://gittar.default.svc.cluster.local",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCenterOrEdgeURL(tt.args.diceCluster, tt.args.requestCluster, tt.args.center, tt.args.saas); got != tt.want {
				t.Errorf("getCenterOrEdgeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
