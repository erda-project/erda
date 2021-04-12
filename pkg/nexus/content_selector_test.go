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
//func TestNexus_ContentSelectorListRequest(t *testing.T) {
//	selectors, err := n.ContentSelectorListRequest(ContentSelectorListRequest{})
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&selectors, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_ContentSelectorCreateRequest(t *testing.T) {
//	err := n.ContentSelectorCreateRequest(ContentSelectorCreateRequest{
//		Name:        "test-content-selector",
//		Description: "test content selector",
//		Expression:  `format == "maven2" and path =^ "/org/sonatype/nexus"`,
//	})
//	assert.NoError(t, err)
//}
//
//func TestNexus_ContentSelectorGetRequest(t *testing.T) {
//	selector, err := n.ContentSelectorGetRequest(ContentSelectorGetRequest{
//		ContentSelectorName: "test-content-selector",
//	})
//	assert.NoError(t, err)
//	s, _ := json.MarshalIndent(&selector, "", "  ")
//	fmt.Println(string(s))
//}
//
//func TestNexus_UpdateContentSelector(t *testing.T) {
//	err := n.UpdateContentSelector(ContentSelectorUpdateRequest{
//		ContentSelectorName: "test-content-selector",
//		Description:         "ssssssss",
//		Expression:          `format == "maven2" and path =^ "/org/sonatype"`,
//	})
//	assert.NoError(t, err)
//}
//
//func TestNexus_DeleteContentSelector(t *testing.T) {
//	err := n.DeleteContentSelector(ContentSelectorDeleteRequest{
//		ContentSelectorName: "test-content-selector",
//	})
//	assert.NoError(t, err)
//}
