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
