package ucauth

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalUSERID(t *testing.T) {
	ui := UserInfo{}
	assert.Nil(t, json.Unmarshal([]byte(`{"id":"dehu"}`), &ui))
	assert.Nil(t, json.Unmarshal([]byte(`{"id": 123}`), &ui))
}
