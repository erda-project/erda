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
