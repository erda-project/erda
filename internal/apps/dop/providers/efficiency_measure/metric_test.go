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

package efficiency_measure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricsItemIDs(t *testing.T) {
	a := []uint64{}
	expectedResult := "0"
	result := GetMetricsItemIDs(a)
	assert.Equal(t, expectedResult, result)

	a = []uint64{1, 2, 3}
	expectedResult = "1 2 3 "
	result = GetMetricsItemIDs(a)
	assert.Equal(t, expectedResult, result)
}
