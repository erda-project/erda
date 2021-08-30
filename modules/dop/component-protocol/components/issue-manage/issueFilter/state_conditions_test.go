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

package issueFilter

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func Test_reorderMemberOption(t *testing.T) {
	type args struct {
		options []filter.PropConditionOption
		userIDs []string
	}

	options := []filter.PropConditionOption{
		{Value: "1"},
		{Value: "2"},
		{Value: "3"},
	}

	tests := []struct {
		name string
		args args
		want []filter.PropConditionOption
	}{
		{
			name: "reorder single item",
			args: args{
				options: options,
				userIDs: []string{
					"2",
				},
			},
			want: []filter.PropConditionOption{
				{Value: "2"},
				{Value: "1"},
				{Value: "3"},
			},
		},
		{
			name: "reorder mutliple items",
			args: args{
				options: options,
				userIDs: []string{
					"2",
					"3",
				},
			},
			want: []filter.PropConditionOption{
				{Value: "2"},
				{Value: "3"},
				{Value: "1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reorderMemberOption(tt.args.options, tt.args.userIDs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reorderMemberOption() = %v, want %v", got, tt.want)
			}
		})
	}
}
