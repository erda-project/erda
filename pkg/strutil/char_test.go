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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsASCII(t *testing.T) {
	assert.True(t, IsASCII('a'))
	assert.True(t, IsASCII('z'))
	assert.True(t, IsASCII('A'))
	assert.True(t, IsASCII('Z'))
	assert.True(t, IsASCII('0'))
	assert.True(t, IsASCII('9'))
	assert.True(t, IsASCII('-'))
	assert.True(t, IsASCII('_'))
	assert.True(t, IsASCII('.'))
	assert.True(t, IsASCII('('))
	assert.True(t, IsASCII('|'))
}

func TestIsAlphaNumericASCII(t *testing.T) {
	assert.True(t, IsAlphaNumericASCII('a'))
	assert.True(t, IsAlphaNumericASCII('z'))
	assert.True(t, IsAlphaNumericASCII('A'))
	assert.True(t, IsAlphaNumericASCII('Z'))
	assert.True(t, IsAlphaNumericASCII('0'))
	assert.True(t, IsAlphaNumericASCII('9'))
	assert.False(t, IsAlphaNumericASCII('-'))
	assert.False(t, IsAlphaNumericASCII('_'))
	assert.False(t, IsAlphaNumericASCII('.'))
	assert.False(t, IsAlphaNumericASCII('('))
	assert.False(t, IsAlphaNumericASCII('|'))
}
