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
