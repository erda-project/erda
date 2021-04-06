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

package diceyml

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestUnmarshalBinds(t *testing.T) {
	s := `
- /aaa:/dddd:ro
- /bbb:/ccc:rw
`
	binds := Binds{}
	assert.Nil(t, yaml.Unmarshal([]byte(s), &binds))

	bindsjson, err := json.Marshal(binds)
	assert.Nil(t, err)
	bindsresult := Binds{}
	assert.Nil(t, json.Unmarshal(bindsjson, &bindsresult))
	assert.Equal(t, binds, bindsresult)
}

func TestUnmarshalVolumes(t *testing.T) {
	s := `
- name~st:/ded/de/de
- /deded
- eee:/ddd
`
	vols := Volumes{}
	assert.Nil(t, yaml.Unmarshal([]byte(s), &vols))

}
