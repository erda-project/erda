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

package apistructs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventCreateRequestMarshal(t *testing.T) {
	eh := EventHeader{
		Event:     "eventname",
		Action:    "eventaction",
		OrgID:     "1",
		ProjectID: "2",
		TimeStamp: "3",
	}
	r := EventCreateRequest{
		EventHeader: eh,
		Sender:      "test",
		Content:     "testcontent",
	}
	m, err := json.Marshal(r)
	assert.Nil(t, err)
	var v interface{}
	assert.Nil(t, json.Unmarshal(m, &v))
	vm := v.(map[string]interface{})
	_, ok := vm["WEBHOOK"]
	assert.False(t, ok)

}
