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

package issueTable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTotalPage(t *testing.T) {
	data := []struct {
		total    uint64
		pageSize uint64
		page     uint64
	}{
		{10, 0, 0},
		{0, 10, 0},
		{20, 10, 2},
		{21, 10, 3},
	}
	for _, v := range data {
		assert.Equal(t, getTotalPage(v.total, v.pageSize), v.page)
	}
}
