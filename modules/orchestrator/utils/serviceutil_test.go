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

func TestRound(t *testing.T) {
	assert.Equal(t, 4.21, Round(4.213213, 2))
	assert.Equal(t, 4.22, Round(4.219213, 2))
}

func TestIsValidK8sSvcName(t *testing.T) {
	assert.True(t, IsValidK8sSvcName("sdf-sdfb---wqfqw-wqfe"))
	assert.True(t, IsValidK8sSvcName("12sdf-sdfb---wqfqw-wqfe213"))
	assert.False(t, IsValidK8sSvcName("sdf-sdfB---wqfqw-wqfe"))
	assert.False(t, IsValidK8sSvcName("sdf-sdfB---wqfqw-wqfe-"))
	assert.False(t, IsValidK8sSvcName(`cmdbcmdbcmdbcmdbcmdbcmdbcmdbcmdbcmdbcmdbcmdbccmdbcmdbcmdb
		cmdbcmdbcmdbcmdbcmdbcmdbcmdbmdbcmdbcmdbcmdbcmdbcmdbcmdbcmdbcmdbcmdb`))
}
