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

package issue

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apps/dop/dao"
)

func Test_importIssueBuilder(t *testing.T) {
	type args struct {
		req       apistructs.Issue
		request   apistructs.IssueImportExcelRequest
		memberMap map[string]string
	}
	tests := []struct {
		name string
		args args
		want dao.Issue
	}{
		{
			args: args{
				req: apistructs.Issue{
					ID:      1,
					Creator: "2",
				},
				request: apistructs.IssueImportExcelRequest{
					IdentityInfo: apistructs.IdentityInfo{
						UserID: "3",
					},
				},
			},
			want: dao.Issue{
				Creator: "3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := importIssueBuilder(tt.args.req, tt.args.request, tt.args.memberMap); !reflect.DeepEqual(got.Creator, tt.want.Creator) {
				t.Errorf("importIssueBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}
