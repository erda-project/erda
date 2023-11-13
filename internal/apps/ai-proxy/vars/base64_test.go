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
	"testing"
)

func TestTryUnwrapBase64(t *testing.T) {
	cases := []string{
		"erda",
		"13012345678",
		"erda@terminus.io",
		"中文",
	}
	for _, c := range cases {
		// base64 encode
		encoded := base64std.EncodeToString([]byte(c))
		// try unwrap
		got := TryUnwrapBase64(encoded)
		if got != c {
			t.Errorf("TryUnwrapBase64(%q) == %q, want %q", encoded, got, c)
		}
		// raw
		got = TryUnwrapBase64(c)
		if got != c {
			t.Errorf("TryUnwrapBase64(%q) == %q, want %q", c, got, c)
		}
	}
}
