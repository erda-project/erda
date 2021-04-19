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

package jsonparse

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/erda-project/erda/pkg/jsonpath"
)

func JsonOneLine(o interface{}) string {
	if o == nil {
		return ""
	}
	switch o.(type) {
	case string: // remove the quotes
		return o.(string)
	case []byte: // remove the quotes
		return string(o.([]byte))
	default:
		var buffer bytes.Buffer
		enc := json.NewEncoder(&buffer)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(o); err != nil {
			return ""
		}
		return strings.TrimSuffix(buffer.String(), "\n")
	}
}

// use jq jsonPath or jackson to filter json
func FilterJson(jsonValue []byte, express string, expressType string) interface{} {
	if len(jsonValue) <= 0 {
		return ""
	}

	var (
		body interface{}
		val  interface{}
	)

	d := json.NewDecoder(bytes.NewReader(jsonValue))
	d.UseNumber()
	err := d.Decode(&body)
	if err != nil {
		return ""
	}

	jsonString := string(jsonValue)

	express = strings.TrimSpace(express)
	if express != "" {
		switch expressType {
		case "jsonpath":
			val, _ = jsonpath.Get(body, express)
		case "jq":
			val, _ = jsonpath.JQ(jsonString, express)
		case "jackson":
			val, _ = jsonpath.Jackson(jsonString, express)
		default:
			if strings.HasPrefix(express, jsonpath.JacksonExpressPrefix) {
				val, _ = jsonpath.Jackson(jsonString, express)
			} else {
				// jq The expression does not necessarily start with ., maybe like '{"ss": 1}' | .ss
				var err error
				val, err = jsonpath.JQ(jsonString, express)
				if err != nil {
					val, _ = jsonpath.Get(body, express)
				}
			}
		}
	} else {
		val = body
	}

	return val
}
