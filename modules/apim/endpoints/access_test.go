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

package endpoints

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, 3},
		{3, 2, 5},
	}
	for _, v := range tests {
		assert.Equal(t, v.want, Add(v.a, v.b))
	}
}
