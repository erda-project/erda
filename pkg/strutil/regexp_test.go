// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
