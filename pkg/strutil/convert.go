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
	"fmt"
	"unsafe"
)

func NoCopyBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func NoCopyStringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func ToStrSlice[T uint64 | uint | int64 | int](input []T, withQuote ...bool) []string {
	result := make([]string, len(input))
	var quote bool
	if len(withQuote) > 0 && withQuote[0] {
		quote = true
	}

	for i, val := range input {
		var str string
		if quote {
			str = fmt.Sprintf("'%v'", val)
		} else {
			str = fmt.Sprintf("%v", val)
		}
		result[i] = str
	}

	return result
}
