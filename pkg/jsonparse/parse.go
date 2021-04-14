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
