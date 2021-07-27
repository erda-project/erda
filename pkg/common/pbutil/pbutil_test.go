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

package pbutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetBool(t *testing.T) {
	// nil
	vv, set := GetBool(nil)
	assert.False(t, vv)
	assert.False(t, set)

	// *true
	vv, set = GetBool(&[]bool{true}[0])
	assert.True(t, vv)
	assert.True(t, set)

	// *false
	vv, set = GetBool(&[]bool{false}[0])
	assert.False(t, vv)
	assert.True(t, set)
}

func TestMustGetBool(t *testing.T) {
	// nil
	vv := MustGetBool(nil)
	assert.False(t, vv)

	// *true
	vv = MustGetBool(&[]bool{true}[0])
	assert.True(t, vv)

	// *false
	vv = MustGetBool(&[]bool{false}[0])
	assert.False(t, vv)
}

func TestGetUint64(t *testing.T) {
	// nil
	vv, set := GetUint64(nil)
	assert.Equal(t, uint64(0), vv)
	assert.False(t, set)

	// 1
	vv, set = GetUint64(&[]uint64{1}[0])
	assert.Equal(t, uint64(1), vv)
	assert.True(t, set)
}

func TestMustGetUint64(t *testing.T) {
	// nil
	vv := MustGetUint64(nil)
	assert.Equal(t, uint64(0), vv)

	// 1
	vv = MustGetUint64(&[]uint64{1}[0])
	assert.Equal(t, uint64(1), vv)
}

func TestGetTimestamp(t *testing.T) {
	// nil *time
	pbt := GetTimestamp(nil)
	assert.Nil(t, pbt)

	// now
	now := time.Now()
	pbt = GetTimestamp(&now)
	assert.NotNil(t, pbt)
	assert.True(t, pbt.AsTime().Equal(now))
}
