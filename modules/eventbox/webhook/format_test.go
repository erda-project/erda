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

package webhook

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractjson(t *testing.T) {
	js := `
{
  "a": {
    "b": 1,
    "ccc": [1,2,3]
  },
  "d": "ddd"
}
`
	var jsv interface{}
	assert.Nil(t, json.Unmarshal([]byte(js), &jsv))
	assert.NotNil(t, extractjson(jsv, []string{"a"}))
	assert.NotNil(t, extractjson(jsv, []string{"a", "b"}))
	assert.NotNil(t, extractjson(jsv, []string{"a", "ccc"}))
	assert.NotNil(t, extractjson(jsv, []string{"d"}))
}
