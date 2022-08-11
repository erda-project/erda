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

package hepautils

import (
	"sort"

	"github.com/erda-project/erda/pkg/strutil"
)

func SortDomains(domains []string) {
	sort.Slice(domains, func(i, j int) bool {
		return strutil.ReverseString(domains[i]) < strutil.ReverseString(domains[j])
	})
}

func SortJoinDomains(domains []string) string {
	SortDomains(domains)
	return strutil.Join(domains, ",")
}
