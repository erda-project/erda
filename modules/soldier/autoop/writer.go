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
