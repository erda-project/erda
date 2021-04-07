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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretVisitor_Visit(t *testing.T) {
	secrets := map[string]string{
		"env_1": "((a))",
		"depth": "1",
		"a":     "b",
	}

	yamlByte := `version: 1.1
stages:
envs:
  ENV_1: ((env_1))
- stage:
  - git-checkout:
      params:
        depth: ((depth))
`

	visitor := NewSecretVisitor([]byte(yamlByte), secrets, 1)
	s := Spec{}
	visitor.Visit(&s)
	assert.Error(t, s.mergeErrors())
}

//func TestRenderSecrets(t *testing.T) {
//	input := []byte("((a))((b))((c))")
//	secret := map[string]string{
//		"a": "1",
//		"b": "2",
//	}
//	output, err := RenderSecrets(input, secret)
//	assert.Error(t, err)
//	_ = output
//
//	secret["c"] = "3"
//	output, err = RenderSecrets(input, secret)
//	assert.NoError(t, err)
//}
