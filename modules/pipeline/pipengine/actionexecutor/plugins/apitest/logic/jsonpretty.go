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

package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
)

// jsonOneLine remove newline added by json encoder.Encode
func jsonOneLine(ctx context.Context, o interface{}) string {
	log := clog(ctx)

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("recover from jsonOneLine: %v", r)
		}
	}()
	if o == nil {
		return ""
	}
	switch o.(type) {
	case string: // 去除引号
		return o.(string)
	case []byte: // 去除引号
		return string(o.([]byte))
	default:
		var buffer bytes.Buffer
		enc := json.NewEncoder(&buffer)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(o); err != nil {
			panic(err)
		}
		return strings.TrimSuffix(buffer.String(), "\n")
	}
}
