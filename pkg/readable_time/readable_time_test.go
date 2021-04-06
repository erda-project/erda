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

package readable_time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadableTime(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2018-12-05T16:54:57+08:00")
	assert.Nil(t, err)
	t2, err := time.Parse(time.RFC3339, "2018-12-05T16:54:59+08:00")
	assert.Nil(t, err)

	a := readableTime(t1, t2)
	assert.Equal(t, int64(2), a.Second)
}
