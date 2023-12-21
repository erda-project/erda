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

package arrays

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsContain(t *testing.T) {
	is := IsContain([]string{"a", "b", "c"}, "a")
	assert.Equal(t, is, true)

}

func TestIsArrayContained(t *testing.T) {
	var index int
	var flag bool
	index, flag = IsArrayContained([]int{1, 3, 4, 5}, []int{1, 3, 4})
	assert.Equal(t, index, -1)
	assert.Equal(t, flag, true)

	index, flag = IsArrayContained([]string{"e", "r", "d", "a"}, []string{"e"})
	assert.Equal(t, index, -1)
	assert.Equal(t, flag, true)

	index, flag = IsArrayContained([]uint64{1}, []uint64{})
	assert.Equal(t, index, -1)
	assert.Equal(t, flag, true)

	index, flag = IsArrayContained([]int64{1, 3, 5, 2}, []int64{2, 5, 4})
	assert.Equal(t, index, 2)
	assert.Equal(t, flag, false)

	index, flag = IsArrayContained([]uint8{}, []uint8{'a'})
	assert.Equal(t, index, 0)
	assert.Equal(t, flag, false)
}
