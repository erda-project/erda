// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
