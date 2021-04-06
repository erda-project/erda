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

package pexpr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindInvalidPlaceholders(t *testing.T) {
	invalidExprs := []string{
		"${{configs.key}}",    // 前后均缺失空格
		"${{ configs.key}}",   // 后面缺失空格
		"${{configs.key }}",   // 前面缺失空格
		"${{  configs.key }}", // 前面有两个空格
		"${{ configs.key  }}", // 后面有两个空格
	}

	for _, expr := range invalidExprs {
		fmt.Println(expr)
		v := FindInvalidPlaceholders(expr)
		assert.True(t, len(v) == 1)
		assert.True(t, v[0] == expr)
	}

	// 正确
	v := FindInvalidPlaceholders("${{ configs.key }}")
	assert.True(t, len(v) == 0)
}
