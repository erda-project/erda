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

package strutil

import (
	"regexp"
	"testing"

	"gotest.tools/assert"
)

func TestReplaceAllStringSubmatchFunc(t *testing.T) {
	s := "${java:OUTPUT:image} ${js:OUTPUT:image}"
	m := map[string]map[string]string{
		"java": {
			"image": "openjdk:8",
		},
		"js": {
			"image": "herd:1.3",
		},
	}
	re := regexp.MustCompile(`\${([^:]+):OUTPUT:([^:]+)}`)
	replaced := ReplaceAllStringSubmatchFunc(re, s, func(sub []string) string {
		return m[sub[1]][sub[2]]
	})
	assert.Equal(t, "openjdk:8 herd:1.3", replaced)
}
