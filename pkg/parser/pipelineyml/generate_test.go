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

package pipelineyml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateYml(t *testing.T) {
	s := &Spec{Version: "1.1"}
	b, err := GenerateYml(s)
	assert.NoError(t, err)
	fmt.Println(string(b))
}

func TestGenerateYml_NilAction(t *testing.T) {
	s := []byte(`
version: 1.1
stages:
- stage:
    - git-checkout:
`)

	y, err := New(s, WithSecrets(map[string]string{}))
	assert.NoError(t, err)

	b, err := GenerateYml(y.s)
	assert.NoError(t, err)
	fmt.Println(string(b))
}
