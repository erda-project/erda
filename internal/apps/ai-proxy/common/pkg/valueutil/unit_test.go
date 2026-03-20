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

package valueutil

import (
	"encoding/json"
	"testing"
)

func TestGetUint64(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		want   uint64
		wantOK bool
	}{
		{name: "float64", input: float64(12), want: 12, wantOK: true},
		{name: "float32", input: float32(13), want: 13, wantOK: true},
		{name: "int", input: int(14), want: 14, wantOK: true},
		{name: "int32", input: int32(15), want: 15, wantOK: true},
		{name: "int64", input: int64(16), want: 16, wantOK: true},
		{name: "uint", input: uint(17), want: 17, wantOK: true},
		{name: "uint32", input: uint32(18), want: 18, wantOK: true},
		{name: "uint64", input: uint64(19), want: 19, wantOK: true},
		{name: "json number", input: json.Number("20"), want: 20, wantOK: true},
		{name: "negative float64", input: float64(-1), want: 0, wantOK: false},
		{name: "negative int", input: int(-2), want: 0, wantOK: false},
		{name: "negative int64", input: int64(-3), want: 0, wantOK: false},
		{name: "invalid json number", input: json.Number("1.5"), want: 0, wantOK: false},
		{name: "string", input: "21", want: 0, wantOK: false},
		{name: "nil", input: nil, want: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetUint64(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if got != tt.want {
				t.Fatalf("expected value=%d, got %d", tt.want, got)
			}
		})
	}
}
