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

package numeral

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormalizeUnitToByte(t *testing.T) {
	ret, err := FormalizeUnitToByte("128974848")
	assert.NoError(t, err)
	assert.Equal(t, int64(128974848), ret)

	ret, err = FormalizeUnitToByte("129K")
	assert.NoError(t, err)
	assert.Equal(t, int64(129000), ret)

	ret, err = FormalizeUnitToByte("123Ki")
	assert.NoError(t, err)
	assert.Equal(t, int64(125952), ret)

	ret, err = FormalizeUnitToByte("129M")
	assert.NoError(t, err)
	assert.Equal(t, int64(129000000), ret)

	ret, err = FormalizeUnitToByte("123Mi")
	assert.NoError(t, err)
	assert.Equal(t, int64(128974848), ret)

	ret, err = FormalizeUnitToByte("129G")
	assert.NoError(t, err)
	assert.Equal(t, int64(129000000000), ret)

	ret, err = FormalizeUnitToByte("123Gi")
	assert.NoError(t, err)
	assert.Equal(t, int64(132070244352), ret)

	ret, err = FormalizeUnitToByte("129T")
	assert.NoError(t, err)
	assert.Equal(t, int64(129000000000000), ret)

	ret, err = FormalizeUnitToByte("123Ti")
	assert.NoError(t, err)
	assert.Equal(t, int64(135239930216448), ret)

	ret, err = FormalizeUnitToByte("129P")
	assert.NoError(t, err)
	assert.Equal(t, int64(129000000000000000), ret)

	ret, err = FormalizeUnitToByte("123Pi")
	assert.NoError(t, err)
	assert.Equal(t, int64(138485688541642752), ret)

	ret, err = FormalizeUnitToByte("1E")
	assert.NoError(t, err)
	assert.Equal(t, int64(1000000000000000000), ret)

	ret, err = FormalizeUnitToByte("1Ei")
	assert.NoError(t, err)
	assert.Equal(t, int64(1152921504606846976), ret)
}
