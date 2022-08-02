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

package testplan

import (
	"reflect"
	"testing"

	issuepb "github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
)

func TestToPbTestPlanCaseRel(t *testing.T) {
	type args struct {
		t apistructs.TestPlanCaseRel
	}
	tests := []struct {
		name string
		args args
		want *issuepb.TestPlanCaseRel
	}{
		{
			args: args{
				t: apistructs.TestPlanCaseRel{
					ID: 1,
				},
			},
			want: &issuepb.TestPlanCaseRel{
				Id: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToPbTestPlanCaseRel(tt.args.t); !reflect.DeepEqual(got.Id, tt.want.Id) {
				t.Errorf("ToPbTestPlanCaseRel() = %v, want %v", got, tt.want)
			}
		})
	}
}
