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

package pointer

import (
	"time"
)

func To[T any](v T) *T {
	return &v
}

func Deref[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}

func DerefPtr[T any](ptr *T, def T) *T {
	if ptr != nil {
		return ptr
	}
	return &def
}

func MustDeref[T any](ptr *T) T {
	if ptr == nil {
		panic("pointer: MustDeref nil pointer")
	}
	return *ptr
}

var String = To[string]
var Int = To[int]
var Int64 = To[int64]
var Uint8 = To[uint8]
var Time = To[time.Time]
var Bool = To[bool]

var StringDeref = Deref[string]
var IntDeref = Deref[int]
var Int64Deref = Deref[int64]
var TimeDeref = Deref[time.Time]
var BoolDeref = Deref[bool]
