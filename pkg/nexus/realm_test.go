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
//func TestNexus_ListActiveRealms(t *testing.T) {
//	users, err := n.ListActiveRealms(RealmListActiveRequest{})
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&users, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_SetActiveRealms(t *testing.T) {
//	err := n.SetActiveRealms(RealmSetActivesRequest{
//		ActiveRealms: []RealmID{
//			NpmBearerTokenRealm.ID,
//		},
//	})
//	assert.NoError(t, err)
//}
//
//func TestNexus_ListAvailableRealms(t *testing.T) {
//	realms, err := n.ListAvailableRealms(RealmListAvailableRequest{})
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&realms, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_EnsureAddRealms(t *testing.T) {
//	assert.NoError(t, n.EnsureAddRealms(RealmEnsureAddRequest{Realms: []RealmID{DockerTokenRealm.ID}}))
//}
