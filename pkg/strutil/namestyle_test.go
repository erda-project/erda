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

package strutil_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/strutil"
)

func TestSnakeToUpCamel(t *testing.T) {
	var names = map[string]string{
		"this_is_a_snake_name": "ThisIsASnakeName",
		"This_Is_A_Snake_name": "ThisIsASnakeName",
	}
	for snake, camel := range names {
		if s := strutil.SnakeToUpCamel(snake); s != camel {
			t.Fatalf("snake: %s, s: %s, camel: %s", snake, snake, camel)
		}
	}
}
