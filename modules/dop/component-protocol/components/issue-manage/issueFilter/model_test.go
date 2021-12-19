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
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func Test_convertConditions(t *testing.T) {
	type args struct {
		status []apistructs.IssueStatus
	}
	tests := []struct {
		name string
		args args
		want []filter.PropConditionOption
	}{
		{
			name: "test",
			args: args{
				status: []apistructs.IssueStatus{
					{
						StateID:   1,
						StateName: "s1",
					},
				},
			},
			want: []filter.PropConditionOption{
				{
					Label: "s1",
					Value: int64(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertConditions(tt.args.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertAllConditions(t *testing.T) {
	type args struct {
		ctx      context.Context
		stateMap map[apistructs.IssueType][]apistructs.IssueStatus
	}
	tests := []struct {
		name string
		args args
		want []filter.PropConditionOption
	}{
		{
			name: "test",
			args: args{
				ctx: nil,
				stateMap: map[apistructs.IssueType][]apistructs.IssueStatus{
					apistructs.IssueTypeRequirement: {
						{
							StateName: "a1",
						},
					},
				},
			},
			want: []filter.PropConditionOption{
				{
					Icon:  "ISSUE_ICON.issue.REQUIREMENT",
					Label: "REQUIREMENT",
					Value: "REQUIREMENT",
					Children: []filter.PropConditionOption{
						{
							Label: "a1",
							Value: int64(0),
						},
					},
				},
			},
		},
	}

	monkey.Patch(cputil.I18n,
		func(ctx context.Context, key string, i ...interface{}) string {
			return "REQUIREMENT"
		})
	defer monkey.UnpatchAll()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertAllConditions(tt.args.ctx, tt.args.stateMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertAllConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}
