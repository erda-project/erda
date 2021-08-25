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

package diceyml

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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
