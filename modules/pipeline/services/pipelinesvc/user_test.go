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

package pipelinesvc

import (
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func TestPipelineSvc_tryGetUser(t *testing.T) {
	bdl := &bundle.Bundle{}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetCurrentUser",
		func(bdl *bundle.Bundle, userID string) (*apistructs.UserInfo, error) {
			return nil, fmt.Errorf("fake error")
		})
	defer m.Unpatch()
	s := &PipelineSvc{bdl: bdl}
	invalidUserID := "invalid user id"
	user := s.tryGetUser(invalidUserID)
	assert.Equal(t, invalidUserID, user.ID)
	assert.Empty(t, user.Name)
}
