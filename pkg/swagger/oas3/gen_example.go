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

package oas3

import (
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
)

// 为 *openapi3.Schema 生成 Example, 就地修改 schema.Example 字段
// 生成前 Schema 应当是展开的, 不能包含引用
func GenExample(schema *openapi3.Schema) {
	genExample(schema, true)
}

func genExample(schema *openapi3.Schema, convStr bool) {
	if schema.Example != nil {
		return
	}

	switch schema.Type {
	case "boolean":
		schema.Example = true
	case "string":
		schema.Example = ""
	case "number":
		schema.Example = 0
	case "object":
		var m = make(map[string]interface{}, 0)
		for key, property := range schema.Properties {
			if property.Value == nil {
				continue
			}
			genExample(property.Value, false)
			m[key] = property.Value.Example
		}
		data, _ := json.MarshalIndent(m, "", "  ")
		if convStr {
			schema.Example = string(data)
		} else {
			schema.Example = json.RawMessage(data)
		}
	case "array":
		var li []interface{}
		if schema.Items == nil || schema.Items.Value == nil {
			schema.Example = json.RawMessage("[]")
			return
		}
		genExample(schema.Items.Value, false)
		li = append(li, schema.Items.Value.Example)
		data, _ := json.MarshalIndent(li, "", "")
		if convStr {
			schema.Example = string(data)
		} else {
			schema.Example = json.RawMessage(data)
		}
	}
}
