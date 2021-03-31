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

var pipeline = `{
    "action": "B_CREATE",
    "applicationID": "178",
    "env": "",
    "event": "pipeline",
    "orgID": "2",
    "projectID": "70",
    "timestamp": "2019-03-13 10:43:49"
}`

// func TestFormat(t *testing.T) {
// 	var pipelinejs interface{}
// 	assert.Nil(t, json.Unmarshal([]byte(pipeline), &pipelinejs))
// 	_, err := Format(pipelinejs)
// 	assert.Nil(t, err)
// }
