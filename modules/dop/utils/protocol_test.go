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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestGetProtocol(t *testing.T) {
	tests := []struct {
		env      string
		expected string
	}{
		{"http", "http"},
		{"https", "https"},
		{"https,http", "https"},
		{"http,https", "http"},
		{"", "https"},
	}
	for _, v := range tests {
		os.Setenv(string(apistructs.DICE_PROTOCOL), v.env)
		assert.Equal(t, v.expected, GetProtocol())
	}

}
