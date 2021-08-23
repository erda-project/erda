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

package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbsInt(t *testing.T) {
	x := AbsInt(10)
	assert.Equal(t, 10, x)

	x = AbsInt(0)
	assert.Equal(t, 0, x)

	x = AbsInt(-10)
	assert.Equal(t, 10, x)
}

func TestAbsInt32(t *testing.T) {
	x := AbsInt32(10)
	assert.Equal(t, int32(10), x)

	x = AbsInt32(0)
	assert.Equal(t, int32(0), x)

	x = AbsInt32(-10)
	assert.Equal(t, int32(10), x)
}

func TestAbsInt64(t *testing.T) {
	x := AbsInt64(10)
	assert.Equal(t, int64(10), x)

	x = AbsInt64(0)
	assert.Equal(t, int64(0), x)

	x = AbsInt64(-10)
	assert.Equal(t, int64(10), x)
}
