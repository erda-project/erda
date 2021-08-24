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
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestJsonOutput_Write(t *testing.T) {
	buf := new(bytes.Buffer)
	data := []byte(`{"node":"1.2.3.4","body":"body","stream":"stderr"}\n`)
	jo := &jsonOutput{w: buf}
	_, err := jo.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, buf.Bytes()) {
		t.Fatal("data not equal")
	}
	if len(jo.a) != 1 && jo.a[0] != (apistructs.AutoopOutputLine{Stream: "stderr", Body: "body", Node: "1.2.3.4"}) {
		t.Fatal("line not equal")
	}
}
