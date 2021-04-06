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

package table

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableNormal(t *testing.T) {
	assert.Nil(t, NewTable().Header([]string{"Header1", "Header2"}).
		Data([][]string{{"D1-1", "D1-2"}, {"D2-1", "D2-2"}}).
		Flush())
}

func TestTableOnlyHeader(t *testing.T) {
	assert.Nil(t, NewTable().Header([]string{"aaa", "bbb"}).Flush())
}

func TestTableEmptyStr(t *testing.T) {
	var buf strings.Builder
	assert.Nil(t, NewTable(WithWriter(&buf)).Header([]string{"", "bb"}).Flush())
	assert.True(t, strings.Contains(buf.String(), "<NIL>"))

	var buf2 strings.Builder
	assert.Nil(t, NewTable(WithWriter(&buf2)).Data([][]string{{"", "bb"}, {"", ""}}).Flush())
	assert.Equal(t, 3, strings.Count(buf2.String(), "<nil>"))
}

func TestTableLongData(t *testing.T) {
	NewTable().Header([]string{"h1", "h2", "h3"}).
		Data([][]string{{"short", "long-long-long-long-long", "short"}}).Flush()
}

func TestVerticalTable(t *testing.T) {
	NewTable(WithVertical()).Header([]string{"h1hhh", "h2hhh", "hhh3"}).
		Data([][]string{{"short", "long-long-long-long-long", "short"}}).Flush()

}

func TestOnlyData(t *testing.T) {
	assert.Nil(t, NewTable(WithVertical()).Data([][]string{{"d1", "d2"}}).Flush())
}
