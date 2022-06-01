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

package persist

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/modules/tools/monitor/core/metric"
)

func Test_getStatisticsLabels(t *testing.T) {
	type args struct {
		data *metric.Metric
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "normal",
			args: args{
				data: &metric.Metric{
					Tags: map[string]string{"org_name": "xxx", "cluster_name": "yyy"},
				},
			},
			want: []string{"xxx", "yyy"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStatisticsLabels(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStatisticsLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
