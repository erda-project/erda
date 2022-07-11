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

package httpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSingleton(t *testing.T) {
	// NewSingleton not invoked
	assert.Nil(t, singleton)

	// NewSingleton invoked
	s1 := NewSingleton("")
	assert.NotNil(t, singleton)
	assert.NotNil(t, s1)
	assert.Equal(t, s1, singleton)

	// NewSingleton invoked again
	s2 := NewSingleton("")
	assert.NotNil(t, singleton)
	assert.NotNil(t, s2)
	assert.Equal(t, s1, singleton)
	assert.Equal(t, s2, s1)
}
