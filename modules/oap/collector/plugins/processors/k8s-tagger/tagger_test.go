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

package tagger

import (
	"testing"

	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/k8s-tagger/metadata/pod"
)

func Test_generateIndexByMatcher(t *testing.T) {
	type args struct {
		matcher string
		tags    map[string]string
	}
	tests := []struct {
		name string
		args args
		want pod.Key
	}{
		{
			args: args{
				matcher: "%{namespace}/%{pod}",
				tags: map[string]string{
					"pod":       "aaa",
					"namespace": "default",
				},
			},
			want: "default/aaa",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateIndexByMatcher(tt.args.matcher, tt.args.tags); got != tt.want {
				t.Errorf("generateIndexByMatcher() = %v, want %v", got, tt.want)
			}
		})
	}
}
