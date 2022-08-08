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

package hepautils_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/hepautils"
	"github.com/erda-project/erda/pkg/strutil"
)

func TestSortDomains(t *testing.T) {
	var domains = getCase()
	var sorted = sortDomains(domains)
	hepautils.SortDomains(domains)
	t.Log(domains)
	t.Log(sorted)
	for i := range domains {
		if domains[i] != sorted[i] {
			t.Fatalf("domains[%v] != sorted[%v]", i, i)
		}
	}
}

func TestSortJoinDomains(t *testing.T) {
	var (
		domains = getCase()
		sorted  = sortDomains(domains)
	)

	s1 := hepautils.SortJoinDomains(domains)
	s2 := strings.Join(sorted, ",")
	t.Log(s1)
	t.Log(s2)
	if s1 != s2 {
		t.Fatalf("s1 != s2")
	}
}

func getCase() []string {
	return []string{
		"dev-gateway.app.terminus.io",
		"dev-gateway.inner",
		"dev-gateway.baidu.com",
		"custom-gateway.app.terminus.io",
	}
}

func sortDomains(domains []string) []string {
	var revert = make([]string, len(domains))
	for i := range domains {
		revert[i] = strutil.ReverseString(domains[i])
	}
	sort.Strings(revert)
	for i := range revert {
		revert[i] = strutil.ReverseString(revert[i])
	}
	return revert
}
