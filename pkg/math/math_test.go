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

package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbsInt(t *testing.T) {
	x := AbsInt(10)
	assert.Equal(t, 10, x)

	x = AbsInt(0)
	assert.Equal(t, 0, x)

	x = AbsInt(-10)
	assert.Equal(t, 10, x)
}

func TestAbsInt32(t *testing.T) {
	x := AbsInt32(10)
	assert.Equal(t, int32(10), x)

	x = AbsInt32(0)
	assert.Equal(t, int32(0), x)

	x = AbsInt32(-10)
	assert.Equal(t, int32(10), x)
}

func TestAbsInt64(t *testing.T) {
	x := AbsInt64(10)
	assert.Equal(t, int64(10), x)

	x = AbsInt64(0)
	assert.Equal(t, int64(0), x)

	x = AbsInt64(-10)
	assert.Equal(t, int64(10), x)
}
