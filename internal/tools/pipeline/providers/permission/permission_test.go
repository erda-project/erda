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

package permission

import (
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		internal string
		wantErr  bool
	}{
		{
			name:     "internal-client",
			userID:   "",
			internal: "bundle",
			wantErr:  false,
		},
		{
			name:     "no permission user",
			userID:   "1",
			internal: "",
			wantErr:  true,
		},
		{
			name:     "with permission user",
			userID:   "2",
			internal: "",
			wantErr:  false,
		},
	}
	bdl := bundle.New()
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckPermission", func(_ *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		if req.UserID == "1" {
			return &apistructs.PermissionCheckResponseData{
				Access: false,
			}, nil
		}
		return &apistructs.PermissionCheckResponseData{
			Access: true,
		}, nil
	})
	defer pm1.Unpatch()
	p := &provider{bdl: bdl}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identityInfo := apistructs.IdentityInfo{
				UserID:         tt.userID,
				InternalClient: tt.internal,
			}
			if err := p.Check(identityInfo, &apistructs.PermissionCheckRequest{
				UserID:   tt.userID,
				Scope:    apistructs.OrgScope,
				Resource: apistructs.OrgResource,
				Action:   apistructs.GetAction,
			}); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
