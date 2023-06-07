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

package reverseproxy

import (
	"net/http/httputil"
	"sync"
)

var (
	DefaultBufferPool = NewBufferPool(1024 * 3)
)

type bufferPool struct {
	pool *sync.Pool
}

func NewBufferPool[IntegerType ~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](size IntegerType) httputil.BufferPool {
	return &bufferPool{
		pool: &sync.Pool{
			New: func() any {
				return make([]byte, size)
			},
		},
	}
}

func (bp *bufferPool) Get() []byte {
	return bp.pool.Get().([]byte)
}

func (bp *bufferPool) Put(buf []byte) {
	bp.pool.Put(buf)
}
