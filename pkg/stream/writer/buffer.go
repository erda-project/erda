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

package writer

import (
	"github.com/recallsong/go-utils/errorx"
)

// Buffer .
type Buffer struct {
	w       Writer
	buf     []interface{}
	maxSize int
}

// NewBuffer .
func NewBuffer(w Writer, max int) *Buffer {
	return &Buffer{
		w:       w,
		buf:     make([]interface{}, 0, max),
		maxSize: max,
	}
}

// Write .
func (b *Buffer) Write(data interface{}) error {
	if len(b.buf)+1 > b.maxSize {
		err := b.Flush()
		if err != nil {
			return err
		}
	}
	b.buf = append(b.buf, data)
	return nil
}

// WriteN 返回 data 已写入 buffer 的数量，如果Flush出现错误，该错误也会被返回
func (b *Buffer) WriteN(data ...interface{}) (int, error) {
	alen := len(b.buf)
	blen := len(data)
	if alen+blen < b.maxSize {
		b.buf = append(b.buf, data...)
		return blen, nil
	}
	writes := 0
	if alen >= b.maxSize {
		// never reached
		err := b.Flush()
		if err != nil {
			return 0, nil
		}
	} else if alen > 0 {
		writes = b.maxSize - alen
		b.buf = append(b.buf, data[0:writes]...)
		err := b.Flush()
		if err != nil {
			return writes, err
		}
		data = data[writes:]
		blen -= writes
	}
	for blen > b.maxSize {
		b.buf = append(b.buf, data[0:b.maxSize]...)
		writes += b.maxSize
		err := b.Flush()
		if err != nil {
			return writes, err
		}
		data = data[b.maxSize:]
		blen -= b.maxSize
	}
	if blen > 0 {
		b.buf = append(b.buf, data...)
		writes += blen
	}
	return writes, nil
}

// Flush .
func (b *Buffer) Flush() error {
	l := len(b.buf)
	if l > 0 {
		n, err := b.w.WriteN(b.buf...)
		b.buf = b.buf[0 : l-n]
		if err != nil {
			return err
		}
	}
	return nil
}

// Size .
func (b *Buffer) Size() int {
	return len(b.buf)
}

// Data .
func (b *Buffer) Data() []interface{} {
	return b.buf
}

// Close .
func (b *Buffer) Close() error {
	return errorx.NewMultiError(b.Flush(), b.w.Close()).MaybeUnwrap()
}
