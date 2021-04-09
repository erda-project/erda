// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"testing"
)

func TestTypeOf(t *testing.T) {
	var (
		bv      bool
		iv      int
		sv      string
		bytes   []byte
		slist   []string
		ilist   []int
		empty   interface{}
		unknown struct{}
	)
	if TypeOf(bv) != BoolType {
		t.Error("bool type error", TypeOf(bv))
	}
	if TypeOf(iv) != NumberType {
		t.Error("number type error", TypeOf(iv))
	}
	if TypeOf(sv) != StringType {
		t.Error("string type error", TypeOf(sv))
	}
	if TypeOf(bytes) != StringType {
		t.Error("[]byte type error", TypeOf(bytes))
	}
	if TypeOf(slist) != StringType {
		t.Error("[]string type error", TypeOf(slist))
	}
	if TypeOf(ilist) != NumberType {
		t.Error("[]int type error", TypeOf(ilist))
	}
	if TypeOf(empty) != "" {
		t.Error("nil type error", TypeOf(empty))
	}
	if TypeOf(unknown) != Unknown {
		t.Error("unknown type error", TypeOf(unknown))
	}
}
