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

package qaparser

import (
	"github.com/erda-project/erda/apistructs"
)

type Totals struct {
	*apistructs.TestTotals
}

type Suite struct {
	*apistructs.TestSuite
}

func NewStatuses(pass, skip, failed, err int) map[apistructs.TestStatus]int {
	return map[apistructs.TestStatus]int{
		apistructs.TestStatusPassed:  pass,
		apistructs.TestStatusSkipped: skip,
		apistructs.TestStatusFailed:  failed,
		apistructs.TestStatusError:   err,
	}
}

func (t *Totals) SetStatuses(statuses map[apistructs.TestStatus]int) *Totals {
	t.Statuses = statuses
	return t
}

func (t *Totals) Add(total *apistructs.TestTotals) *Totals {
	t.Tests += total.Tests
	t.Duration += total.Duration
	t.Statuses[apistructs.TestStatusPassed] += total.Statuses[apistructs.TestStatusPassed]
	t.Statuses[apistructs.TestStatusSkipped] += total.Statuses[apistructs.TestStatusSkipped]
	t.Statuses[apistructs.TestStatusFailed] += total.Statuses[apistructs.TestStatusFailed]
	t.Statuses[apistructs.TestStatusError] += total.Statuses[apistructs.TestStatusError]
	return t
}

// Aggregate calculates result sums across all tests.
func (s *Suite) Aggregate() {
	//totals := Totals{Tests: len(s.Tests)}
	totals := &apistructs.TestTotals{
		Tests:    len(s.Tests),
		Statuses: make(map[apistructs.TestStatus]int),
	}

	for _, test := range s.Tests {
		totals.Duration += test.Duration
		switch test.Status {
		case apistructs.TestStatusPassed:
			totals.Statuses[apistructs.TestStatusPassed] += 1
		case apistructs.TestStatusSkipped:
			totals.Statuses[apistructs.TestStatusSkipped] += 1
		case apistructs.TestStatusFailed:
			totals.Statuses[apistructs.TestStatusFailed] += 1
		case apistructs.TestStatusError:
			totals.Statuses[apistructs.TestStatusError] += 1
		}
	}

	s.Totals = totals
}
