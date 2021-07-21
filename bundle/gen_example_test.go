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

package bundle_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/bundle"
)

const schemaA = `description: ""
example:
  str: Example
properties:
  str:
    example: Example
    type: string
    x-dice-name: str
    x-dice-required: true
required:
  - str
type: object`

const schemaAExample = `{
  "str": "Example"
}`

func TestGenExample(t *testing.T) {
	s := openapi3.NewSchema()
	if err := yaml.Unmarshal([]byte(schemaA), s); err != nil {
		t.Fatalf("failed to yaml.Unmarshal schemaA: %v", err)
	}
	bundle.GenExample(s)

	t.Log("GenExample:", s.Example)
	if example, ok := s.Example.(string); !ok {
		t.Fatal("s.Example was not converted to string")
	} else if example != schemaAExample {
		t.Fatal("s.Example was converted wrong")
	}

}
