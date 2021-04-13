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
