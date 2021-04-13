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

package autoop

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/erda-project/erda/apistructs"
)

type jsonLine struct {
	Stream string      `json:"stream"`
	Node   interface{} `json:"node"`
	Host   interface{} `json:"host"`
	Body   string      `json:"body"`
}

type jsonOutput struct {
	w io.Writer
	a []apistructs.AutoopOutputLine
}

func (jo *jsonOutput) Write(p []byte) (int, error) {
	for a := p; len(a) > 0; {
		var b []byte
		if i := bytes.IndexByte(a, '\n'); i == -1 {
			b = a
			a = nil
		} else {
			b = a[:i]
			a = a[i+1:]
		}
		if len(b) > 0 {
			var aol apistructs.AutoopOutputLine
			var jl jsonLine
			if b[0] == '{' && b[len(b)-1] == '}' && json.Unmarshal(b, &jl) == nil {
				aol.Stream = jl.Stream
				aol.Node, _ = jl.Node.(string)
				aol.Host, _ = jl.Host.(string)
				aol.Body = jl.Body
			} else {
				aol.Stream = "stderr"
				aol.Body = string(b)
			}
			jo.a = append(jo.a, aol)
		}
	}
	return jo.w.Write(p)
}
