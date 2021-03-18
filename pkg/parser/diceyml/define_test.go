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
