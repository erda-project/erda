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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidOrgName(t *testing.T) {
	ss := []struct {
		s    string
		want bool
	}{
		{
			"",
			false,
		},
		{
			"a",
			true,
		},
		{
			"1",
			false,
		},
		{
			"1233",
			false,
		},
		{
			"-",
			false,
		},
		{
			"-123",
			false,
		},
		{
			"3434-",
			false,
		},
		{
			"abd-3-c",
			true,
		},
		{
			"123-456",
			true,
		},
		{
			"123a-456-abc",
			true,
		},
		{
			"jfljef",
			true,
		},
		{
			"a---b",
			true,
		},
		{
			"a34jjl",
			true,
		},
		{
			"123a",
			true,
		},
		{
			"1----a",
			true,
		},
	}
	for _, v := range ss {
		assert.Equal(t, v.want, IsValidOrgName(v.s))
	}
}
