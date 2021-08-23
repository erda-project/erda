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

package bundle

import (
	"encoding/json"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// 为 *openapi3.Schema 生成 Example, 就地修改 schema.Example 字段
// 生成前 Schema 应当是展开的, 不能包含引用
func GenExample(schema *openapi3.Schema) {
	genExample(schema, true)
}

func genExample(schema *openapi3.Schema, convStr bool) {
	if schema.Example != nil {
		switch schema.Example.(type) {
		case string:
			return
		case []byte:
			if convStr {
				schema.Example = string(schema.Example.([]byte))
			}
			return
		case json.RawMessage:
			if convStr {
				schema.Example = string(schema.Example.(json.RawMessage))
			}
			return
		}

		data, err := json.MarshalIndent(schema.Example, "", "  ")
		if err != nil {
			return
		}
		schema.Example = string(data)
		return
	}

	switch strings.ToLower(schema.Type) {
	case "boolean":
		schema.Example = true
	case "string":
		schema.Example = ""
	case "number", "int", "integer":
		schema.Example = 0
	case "object":
		var m = make(map[string]interface{})
		for key, property := range schema.Properties {
			if property.Value == nil {
				continue
			}
			genExample(property.Value, false)
			m[key] = property.Value.Example
		}
		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return
		}
		schema.Example = json.RawMessage(data)
		if convStr {
			schema.Example = string(data)
		}
	case "array":
		var li []interface{}
		if schema.Items == nil || schema.Items.Value == nil {
			schema.Example = json.RawMessage("[]")
			if convStr {
				schema.Example = "[]"
			}
			return
		}
		genExample(schema.Items.Value, false)
		li = append(li, schema.Items.Value.Example)
		data, err := json.MarshalIndent(li, "", "")
		if err != nil {
			return
		}
		if convStr {
			schema.Example = string(data)
		} else {
			schema.Example = json.RawMessage(data)
		}
	}
}
