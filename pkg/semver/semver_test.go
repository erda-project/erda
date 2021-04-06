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

package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {
	validVersions := []string{
		"3.4.1",
		"v3.4.1",
		"3.5.0",
		"v3.5.0",
		"3.4.1-fix-your-bug",
		"v3.5.0-fix-123-bug-456",
	}
	for _, ver := range validVersions {
		assert.True(t, Valid(ver), ver)
	}

	invalidVersions := []string{
		"3",
		"v3",
		"3.4",
		"v3.4",
		"3.4.1.1",
		"v3.4.1.1",
		"v3.4.1@",
	}
	for _, ver := range invalidVersions {
		assert.False(t, Valid(ver), ver)
	}
}

func TestNew(t *testing.T) {
	v1 := New(3)
	assert.Equal(t, "3.0.0", v1)

	v2 := New(3, 5)
	assert.Equal(t, "3.5.0", v2)

	v3 := New(3, 5, 1)
	assert.Equal(t, "3.5.1", v3)
}
