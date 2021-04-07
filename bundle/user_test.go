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

package bundle

//import (
//	"os"
//	"testing"
//
//	"github.com/davecgh/go-spew/spew"
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestGetCurrentUser(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	b := New(WithCMDB())
//	userInfo, err := b.GetCurrentUser("2")
//	assert.NoError(t, err)
//	spew.Dump(userInfo)
//}
//
//func TestListUsers(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	b := New(WithCMDB())
//	userInfo, err := b.ListUsers(apistructs.UserListRequest{
//		Query:   "",
//		UserIDs: []string{"1", "2"},
//	})
//	assert.NoError(t, err)
//	spew.Dump(userInfo)
//}
