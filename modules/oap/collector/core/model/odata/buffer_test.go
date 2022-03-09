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

package odata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer_Push_Pop(t *testing.T) {
	ass := assert.New(t)
	buf := NewBuffer(3)
	buf.Push(NewRaw([]byte("1")))
	buf.Push(NewRaw([]byte("2")))
	buf.Push(NewRaw([]byte("3")))
	ass.True(buf.Full())
	buf.Push(NewRaw([]byte("4")))

	ass.Equal("1", string(buf.Pop().Source().([]byte)))
	ass.Equal("2", string(buf.Pop().Source().([]byte)))
	ass.Equal("3", string(buf.Pop().Source().([]byte)))
	ass.True(buf.Empty())
	ass.Empty(buf.FlushAll())
}
