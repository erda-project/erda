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

package websocket

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

const (
	// Frame header byte 1 bits from Section 5.2 of RFC 6455
	maskBit = 1 << 7
)

func DecodeFrame(data []byte) []byte {
	buf := bytes.NewBuffer(data)
	first2Bytes := make([]byte, 2)
	buf.Read(first2Bytes)

	mask := first2Bytes[1]&maskBit != 0
	length := uint64(first2Bytes[1] & 0x7f)
	switch length {
	case 126:
		lenBytes := make([]byte, 2)
		buf.Read(lenBytes)
		length = binary.BigEndian.Uint64(lenBytes)
	case 127:
		lenBytes := make([]byte, 8)
		buf.Read(lenBytes)
		length = binary.BigEndian.Uint64(lenBytes)
	}

	maskKeyBytes := make([]byte, 4)
	if mask {
		buf.Read(maskKeyBytes)
	}

	payloadBytes := make([]byte, length)
	buf.Read(payloadBytes)
	if mask {
		maskBytes(maskKeyBytes, 0, payloadBytes)
	}
	return payloadBytes
}

const wordSize = int(unsafe.Sizeof(uintptr(0)))

func maskBytes(key []byte, pos int, b []byte) int {
	// Mask one byte at a time for small buffers.
	if len(b) < 2*wordSize {
		for i := range b {
			b[i] ^= key[pos&3]
			pos++
		}
		return pos & 3
	}

	// Mask one byte at a time to word boundary.
	if n := int(uintptr(unsafe.Pointer(&b[0]))) % wordSize; n != 0 {
		n = wordSize - n
		for i := range b[:n] {
			b[i] ^= key[pos&3]
			pos++
		}
		b = b[n:]
	}

	// Create aligned word size key.
	var k [wordSize]byte
	for i := range k {
		k[i] = key[(pos+i)&3]
	}
	kw := *(*uintptr)(unsafe.Pointer(&k))

	// Mask one word at a time.
	n := (len(b) / wordSize) * wordSize
	for i := 0; i < n; i += wordSize {
		*(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&b[0])) + uintptr(i))) ^= kw
	}

	// Mask one byte at a time for remaining bytes.
	b = b[n:]
	for i := range b {
		b[i] ^= key[pos&3]
		pos++
	}

	return pos & 3
}
