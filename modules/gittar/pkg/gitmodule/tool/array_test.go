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

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsKeyInArray(t *testing.T) {
	tests := []struct {
		array    []string
		key      string
		expected bool
	}{
		{[]string{"foo", "bar"}, "foo", true},
		{[]string{"foo", "bar"}, "", false},
		{[]string{"foo", ""}, "", true},
	}
	for _, v := range tests {
		assert.Equal(t, v.expected, IsKeyInArray(v.array, v.key))
	}
}
