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
