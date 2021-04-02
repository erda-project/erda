package errors

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackendErrMarshal(t *testing.T) {
	var b BackendErrs = make(map[string][]error)
	b["asd"] = []error{errors.New("EEE")}
	v, err := json.Marshal(&b)
	assert.Nil(t, err)
	assert.Equal(t, "{\"asd\":[\"EEE\"]}", string(v))
}

func TestMarshal(t *testing.T) {
	d := New()
	d.BackendErrs = map[string][]error{
		"AA": {errors.New("EEE")},
	}
	v, err := json.Marshal(d)
	assert.Nil(t, err)
	assert.Equal(t, "{\"BackendErrs\":{\"AA\":[\"EEE\"]},\"FilterInfo\":\"\",\"FilterErr\":\"\"}", string(v))
}
