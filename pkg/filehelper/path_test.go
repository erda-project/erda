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
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestAbs2Rel(t *testing.T) {
	path1 := "/"
	path1 = Abs2Rel(path1)
	assert.Equal(t, ".", path1)

	path2 := "//testdata/"
	path2 = Abs2Rel(path2)
	assert.Equal(t, "testdata", path2)
}

func TestAPIFileUrlRetriever(t *testing.T) {
	path1 := "https://test.io/api/files/xed"
	path1 = APIFileUrlRetriever(path1)
	assert.Equal(t, "/api/files/xed", path1)
	path2 := "api/files/xxx"
	path2 = APIFileUrlRetriever(path2)
	assert.Equal(t, "api/files/xxx", path2)
	path3 := "https://test.io/erda/api/files/xxx"
	path3 = APIFileUrlRetriever(path3)
	assert.Equal(t, "https://test.io/erda/api/files/xxx", path3)
}

func TestFilterAPIFileUrl(t *testing.T) {
	path1 := "(dasd)![a.png(94.56 KB)](https://test.io/api/files/xed)![b.png(94.56 KB)](https://test.io/api/files/xe2)"
	path1 = FilterAPIFileUrl(path1)
	assert.Equal(t, "(dasd)![a.png(94.56 KB)](/api/files/xed)![b.png(94.56 KB)](/api/files/xe2)", path1)
	path2 := "(/api/files/xev)"
	path2 = FilterAPIFileUrl(path2)
	assert.Equal(t, "(/api/files/xev)", path2)
}
