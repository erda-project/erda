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

package filehelper

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateFile(t *testing.T) {
	err := CreateFile("/tmp/ci/test.sh", "echo hello\necho world\necho nice", 0700)
	require.NoError(t, err)
}

func TestCreateFile2(t *testing.T) {
	err := CreateFile2("/tmp/ci/test.sh", bytes.NewBufferString("ssssxfdsfs\nfs"), 0700)
	require.NoError(t, err)
}
