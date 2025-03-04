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
	"testing"
)

func TestStripANSIEscapes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "No escape sequences",
			input:    []byte("Hello, World!"),
			expected: []byte("Hello, World!"),
		},
		{
			name:     "With escape sequences",
			input:    []byte("\u001B[31mHello, World!\u001B[0m"),
			expected: []byte("Hello, World!"),
		},
		{
			name:     "Mixed content",
			input:    []byte("Normal \u001B[31mRed\u001B[0m Text"),
			expected: []byte("Normal Red Text"),
		},
		{
			name:     "Empty input",
			input:    []byte(""),
			expected: []byte(""),
		},
		{
			name:     "With color matched",
			input:    []byte(`MYSQLX_[01;31m[KPAS[m[KSWORD,ALIYUN_[01;31m[KSECRET[m[K_KEY`),
			expected: []byte(`MYSQLX_PASSWORD,ALIYUN_SECRET_KEY`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripANSIEscapes(tt.input)
			if string(result) != string(tt.expected) {
				t.Errorf("AnsiEscape() = %v, want %v", string(result), string(tt.expected))
			}
			stringResult := StripANSIEscapesString(string(tt.input))
			if stringResult != string(tt.expected) {
				t.Errorf("AnsiEscape() = %v, want %v", string(result), string(tt.expected))
			}
		})
	}
}
