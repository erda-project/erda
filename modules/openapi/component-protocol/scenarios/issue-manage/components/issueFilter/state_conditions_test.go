// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
