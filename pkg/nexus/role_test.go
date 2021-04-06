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

package nexus

//import (
//	"encoding/json"
//	"fmt"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestNexus_ListRoles(t *testing.T) {
//	users, err := n.ListRoles(RoleListRequest{
//		Source: "default",
//	})
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&users, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_CreateRole(t *testing.T) {
//	id := "sdk-npm-hosted-100-deployment"
//
//	err := n.CreateRole(RoleCreateRequest{
//		ID:          RoleID(id),
//		Name:        id,
//		Description: "desc",
//		Privileges: []PrivilegeID{
//			"nx-repository-view-npm-sdk-npm-hosted-100-add",
//			"nx-repository-view-npm-sdk-npm-hosted-100-browse",
//			"nx-repository-view-npm-sdk-npm-hosted-100-edit",
//			"nx-repository-view-npm-sdk-npm-hosted-100-read",
//		},
//		Roles: []string{},
//	})
//	assert.NoError(t, err)
//
//	role, err := n.GetRole(RoleGetRequest{
//		ID: RoleID(id),
//	})
//	assert.NoError(t, err)
//	printJSON(role)
//}
//
//func TestNexus_UpdateRole(t *testing.T) {
//	id := "b489a2f09421431eb8d49f83d8b81c20"
//
//	err := n.UpdateRole(RoleUpdateRequest{
//		ID:          RoleID(id),
//		Name:        id,
//		Description: "desc",
//		Privileges:  []PrivilegeID{"test-content-selector-privilege-npm"},
//		Roles:       []string{"id2"},
//	})
//	assert.NoError(t, err)
//
//	role, err := n.GetRole(RoleGetRequest{
//		ID: RoleID(id),
//	})
//	assert.NoError(t, err)
//	printJSON(role)
//}
//
//func TestNexus_GetRole(t *testing.T) {
//	role, err := n.GetRole(RoleGetRequest{
//		ID: "id2",
//	})
//	assert.NoError(t, err)
//	printJSON(role)
//}
//
//func TestNexus_DeleteRole(t *testing.T) {
//	err := n.DeleteRole(RoleDeleteRequest{
//		ID: "id2",
//	})
//	assert.NoError(t, err)
//}
