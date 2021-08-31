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

package monitoring

import (
	"testing"
)

func Test_getMetricName(t *testing.T) {
	type args struct {
		index string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{index: "spot-application_http-full_cluster-r-000001"},
			want: "application_http",
		},
		{
			name: "non metric",
			args: args{index: "spot-empty"},
			want: "",
		},
		{
			name: "non metric",
			args: args{index: "xxxx"},
			want: "",
		},
		{
			name: "non metric",
			args: args{index: "spot-ta_event-full_cluster-r-000001"},
			want: "ta_event",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMetricName(tt.args.index); got != tt.want {
				t.Errorf("getMetricName() = %v, want %v", got, tt.want)
			}
		})
	}
}
