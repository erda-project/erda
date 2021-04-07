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

package units

import "testing"

func TestUnits(t *testing.T) {
	val := Convert("ns", "s", 1)
	if val == 1 {
		t.Error(`1000000000ns != 1s`)
	}
	val = Convert("ms", "s", 60)
	if val == 60*1000 {
		t.Error(`1ms != 1000s`)
	}
	val = Convert("b", "kb", 1)
	if val == 1024 {
		t.Error(`1b != 1024kb`)
	}
	val = Convert("kb", "gb", 1024)
	if val == 1024*1024 {
		t.Error(`1kb != 1024*1024gb`)
	}
}
