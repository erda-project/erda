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
	"reflect"
	"testing"
)

func TestToStrSlice(t *testing.T) {
	testCases := []struct {
		input     interface{}
		withQuote bool
		expected  []string
	}{
		{
			input:     []int{1, 2, 3, 4, 5},
			withQuote: true,
			expected:  []string{"'1'", "'2'", "'3'", "'4'", "'5'"},
		},
		{
			input:    []int64{10, 20, 30, 40, 50},
			expected: []string{"10", "20", "30", "40", "50"},
		},
		{
			input:     []uint{100, 200, 300, 400, 500},
			withQuote: true,
			expected:  []string{"'100'", "'200'", "'300'", "'400'", "'500'"},
		},
		{
			input:    []uint64{1000, 2000, 3000, 4000, 5000},
			expected: []string{"1000", "2000", "3000", "4000", "5000"},
		},
	}

	for _, testCase := range testCases {
		var result []string
		switch input := testCase.input.(type) {
		case []int:
			result = ToStrSlice(input, testCase.withQuote)
		case []int64:
			result = ToStrSlice(input, testCase.withQuote)
		case []uint:
			result = ToStrSlice(input, testCase.withQuote)
		case []uint64:
			result = ToStrSlice(input, testCase.withQuote)
		default:
			t.Errorf("Unsupported input type: %T", testCase.input)
			continue
		}

		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("Input: %v, withQuote: %v\nExpected: %v\nGot: %v", testCase.input, testCase.withQuote, testCase.expected, result)
		}
	}
}
