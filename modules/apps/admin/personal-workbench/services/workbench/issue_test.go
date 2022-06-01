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

package workbench

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func TestWorkbench_GetIssueQueries(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		wantErr  bool
		stateErr bool
	}{
		{
			name:     "dop_issue_state_query_error",
			wantErr:  false,
			stateErr: true,
		},
		{
			name:     "query_success",
			wantErr:  false,
			stateErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projIDs := []uint64{12, 13, 16}

			identity := apistructs.Identity{
				UserID: "2",
				OrgID:  "1",
			}
			bdl := &bundle.Bundle{}
			wb := New(WithBundle(bdl))
			_, err := wb.GetProjIssueQueries(identity.UserID, projIDs, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjIssueQueries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
