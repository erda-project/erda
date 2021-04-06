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

package gitflowutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsReleaseTag(t *testing.T) {
	validTags := []string{
		"3.4.1",
		"v3.4.1",
		"3.5.0",
		"v3.5.0",
		"3.4.1-fix-your-bug",
		"v3.5.0-fix-123-bug-456",
	}
	for _, tag := range validTags {
		assert.True(t, IsReleaseTag(tag), tag)
	}

	invalidTags := []string{
		"3",
		"v3",
		"3.4",
		"v3.4",
		"3.4.1.1",
		"v3.4.1.1",
		"v3.4.1@",
	}
	for _, tag := range invalidTags {
		assert.False(t, IsReleaseTag(tag), tag)
	}
}
