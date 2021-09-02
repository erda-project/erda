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

package member

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/dao"
)

func Test_checkCreateParam(t *testing.T) {
	m := New()
	req := apistructs.MemberAddRequest{
		Roles: []string{"Auditor"},
		Scope: apistructs.Scope{
			Type: "sys",
			ID:   "0",
		},
		UserIDs: []string{"2", "3"},
	}
	err := m.checkCreateParam(req)
	assert.NoError(t, err)
}

func Test_CheckPermission(t *testing.T) {
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "IsSysAdmin",
		func(_ *dao.DBClient, userID string) (bool, error) {
			return userID == "1", nil
		})
	defer monkey.UnpatchAll()
	m := New()
	m.db = db
	err := m.CheckPermission("1", apistructs.SysScope, 0)
	assert.NoError(t, err)
}
