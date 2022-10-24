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

package common

import (
	"reflect"
	"testing"
)

func Test_ParsePublicHostsFromLabel(t *testing.T) {
	type args struct {
		labels map[string]string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "none",
			args: args{
				labels: map[string]string{},
			},
			want: []string{},
		},
		{
			name: "parse",
			args: args{
				labels: map[string]string{
					LabelHAProxyVHost: "test1.erda.cloud,test2.erda.cloud",
				},
			},
			want: []string{
				"test1.erda.cloud",
				"test2.erda.cloud",
			},
		},
		{
			name: "single",
			args: args{
				labels: map[string]string{
					LabelHAProxyVHost: "test.erda.cloud",
				},
			},
			want: []string{
				"test.erda.cloud",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParsePublicHostsFromLabel(tt.args.labels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePublicHostsFromLabel got %v, want %v", got, tt.want)
			}
		})
	}
}
