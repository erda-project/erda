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
