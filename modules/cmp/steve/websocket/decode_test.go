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

package websocket

import (
	"encoding/base64"
	"testing"
)

func TestDecodeFrame(t *testing.T) {
	frame := []byte{129, 144, 155, 182, 186, 247, 250, 241, 236, 132, 249, 241, 130, 144, 255, 132, 131, 142, 249,
		241, 235, 202, 0, 62, 249, 122, 86, 4, 248, 80, 11, 4, 165, 72, 17, 5, 140, 87, 44,
	}
	data := DecodeFrame(frame)
	res, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		t.Error(err)
	}
	if string(res) != "hello world" {
		t.Errorf("test failed, expect res %s, actual %s", "hello world", string(res))
	}
}
