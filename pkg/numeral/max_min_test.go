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

func TestMinFloat64(t *testing.T) {
	min := MinFloat64([]float64{1, 2, 3, 4, 5, 0, 0})
	assert.Equal(t, min, float64(0))

	min = MinFloat64([]float64{1, 2, 3, 4, 5, 0, 0}, true)
	assert.Equal(t, min, float64(1))

	min = MinFloat64([]float64{-1.1, -2.3, 0, 1.2, 100})
	assert.Equal(t, min, -2.3)

	min = MinFloat64([]float64{0, 1, 2, 3, 4}, true)
	assert.Equal(t, min, float64(1))
}

func TestMaxFloat64(t *testing.T) {
	max := MaxFloat64([]float64{1, 2, 3, 4, 5})
	assert.Equal(t, max, float64(5))

	max = MaxFloat64([]float64{-2, -3, -100, -0.5, 0})
	assert.Equal(t, max, float64(0))
}
