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

package nodeFilter

import (
	"reflect"
	"testing"

	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/filter"
)

func TestDoFilter(t *testing.T) {
	type args struct {
		nodeList []data.Object
		values   filter.Values
	}
	d := []data.Object{{"id": "nameID", "metadata": map[string]interface{}{"name": "name", "labels": map[string]interface{}{"key1": "value1"}}}}
	tests := []struct {
		name string
		args args
		want []data.Object
	}{
		{
			name: "test",
			args: args{
				nodeList: d,
				values:   map[string]interface{}{"Q": "na", "service": []interface{}{"key1=value1", "serviceLabel2"}},
			},
			want: d,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DoFilter(tt.args.nodeList, tt.args.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DoFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
