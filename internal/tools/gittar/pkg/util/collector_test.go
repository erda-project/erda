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

package util

import (
	"testing"

	"github.com/erda-project/erda/internal/tools/gittar/metrics"
)

func TestIsValidBranch(t *testing.T) {
	testCases := []struct {
		s        string
		prefixes []string
		expected bool
	}{
		{"feature/branch", []string{"feature/"}, true},
		{"bugfix/branch", []string{"feature/", "bugfix/"}, true},
		{"main", []string{"feature/"}, false},
		{"release/v1.0", []string{"release/"}, true},
	}

	for _, tc := range testCases {
		result := metrics.IsValidBranch(tc.s, tc.prefixes...)
		if result != tc.expected {
			t.Errorf("isValidBranch(%s, %v) = %v, expected %v", tc.s, tc.prefixes, result, tc.expected)
		}
	}
}
