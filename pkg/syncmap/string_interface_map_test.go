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

package syncmap

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestStringInterfaceMap_MarshalJSON(t *testing.T) {
	jstr := `{"1":{"name":"name1","value":"value1"},"2":2}`
	var m StringInterfaceMap
	m.Store("1", apistructs.MetadataField{Name: "name1", Value: "value1"})
	m.Store("2", 2)
	b, err := json.Marshal(&m)
	assert.NoError(t, err)
	assert.Equal(t, jstr, string(b))
}

func TestStringInterfaceMap_UnmarshalJSON(t *testing.T) {
	jstr := `{"1":{"name":"name1","value":"value1"},"2":2}`
	var m StringInterfaceMap
	err := json.Unmarshal([]byte(jstr), &m)
	assert.NoError(t, err)

	var o apistructs.MetadataField
	err = m.Get("1", &o)
	assert.NoError(t, err)
	assert.Equal(t, "name1", o.Name)
	assert.Equal(t, "value1", o.Value)

	var i int
	err = m.Get("2", &i)
	assert.NoError(t, err)
	assert.Equal(t, 2, i)
}

func TestStringInterfaceMap_MarshalJSON2(t *testing.T) {
	type extra struct {
		Name    string             `json:"name"`
		Volumes StringInterfaceMap `json:"volumes"`
	}
	var e extra
	e.Name = "name"
	e.Volumes = StringInterfaceMap{}
	e.Volumes.Store("k1", apistructs.MetadataField{Name: "n1"})
	e.Volumes.Store("k2", apistructs.MetadataField{Name: "n2"})

	b, err := json.Marshal(&e)
	assert.NoError(t, err)
	fmt.Println(string(b))
}

func TestStringInterfaceMap_GetMap(t *testing.T) {
	var m StringInterfaceMap
	m.Store("1", apistructs.MetadataField{Name: "name1", Value: "value1"})
	m.Store("2", 2)
	tmpMap := m.GetMap()
	assert.Equal(t, 2, len(tmpMap))
	assert.Equal(t, tmpMap["1"], apistructs.MetadataField{Name: "name1", Value: "value1"})
	assert.Equal(t, tmpMap["2"], 2)
}
