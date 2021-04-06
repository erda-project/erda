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
