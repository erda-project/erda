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
	orgName1 := "123"
	orgName2 := "123-a"
	orgName3 := ""
	orgName4 := "a-1"
	orgName5 := "12345677890"
	orgName6 := "12df-fjel"
	orgName7 := "-123"
	orgName8 := "123-"
	orgName9 := "*fjle"
	orgName10 := "fjlejf*"
	assert.Equal(t, IsValidOrgName(orgName1), false)
	assert.Equal(t, IsValidOrgName(orgName2), true)
	assert.Equal(t, IsValidOrgName(orgName3), false)
	assert.Equal(t, IsValidOrgName(orgName4), true)
	assert.Equal(t, IsValidOrgName(orgName5), false)
	assert.Equal(t, IsValidOrgName(orgName6), true)
	assert.Equal(t, IsValidOrgName(orgName7), false)
	assert.Equal(t, IsValidOrgName(orgName8), false)
	assert.Equal(t, IsValidOrgName(orgName9), false)
	assert.Equal(t, IsValidOrgName(orgName10), false)
}
