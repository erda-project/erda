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

package vars

import (
	"encoding/base64"
	"unicode/utf8"
)

var base64std = base64.StdEncoding

func TryUnwrapBase64(v string) string {
	decoded, err := base64std.DecodeString(v)
	if err == nil {
		// if decoded value contains non-utf-8 characters, treat it as raw value
		if !utf8.Valid(decoded) {
			return v
		}
		return string(decoded)
	}
	return v
}
