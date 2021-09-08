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

	"bou.ke/monkey"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
)

func TestTestPlan_createAudit(t *testing.T) {
	type args struct {
		testPlan *dao.TestPlan
		req      apistructs.TestPlanUpdateRequest
	}
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProject",
		func(bdl *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateAuditEvent",
		func(bdl *bundle.Bundle, audits *apistructs.AuditCreateRequest) error {
			return nil
		},
	)
	defer monkey.UnpatchAll()

	tr := New(WithBundle(bdl))
	var archive = true
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				testPlan: &dao.TestPlan{},
				req: apistructs.TestPlanUpdateRequest{
					IsArchived: &archive,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tr.createAudit(tt.args.testPlan, tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("TestPlan.createAudit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
