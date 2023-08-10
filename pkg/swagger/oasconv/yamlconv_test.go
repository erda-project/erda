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

package oasconv_test

import (
	"encoding/json"
	"testing"
	
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/pkg/swagger/oasconv"
)

const y = `$schema: http://json-schema.org/draft-07/schema#
type: object
properties:
  name:
    type: string
    description: the name of the case
  preCondition:
    type: string
    description: pre condition
    items:
      type: string
  stepAndResults:
    type: array
    description: list of step and corresponding result
    items:
      type: object
      properties:
        step:
          type: string
          description: Operation steps
        result:
          type: string
          description: Expected result
required:
  - name
  - preCondition
  - stepAndResults
`

func BenchmarkYAMLToJSON(b *testing.B) {
	b.Run("YAMLToJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := oasconv.YAMLToJSON([]byte(y)); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("YAMLToJSON by unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := YAMLToJSON([]byte(y)); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func YAMLToJSON(data []byte) ([]byte, error) {
	var j = make(json.RawMessage, 0)
	err := yaml.Unmarshal(data, &j)
	return j, err
}
