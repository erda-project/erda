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

	"github.com/stretchr/testify/assert"
)

func Test_orgNameRetriever(t *testing.T) {
	var domains = []string{"erda-org.erda.cloud", "buzz-org.app.terminus.io", "fuzz.com"}
	var domainRoots = []string{"erda.cloud", "app.terminus.io"}
	assert.Equal(t, "erda", orgNameRetriever(domains[0], domainRoots[0]))
	assert.Equal(t, "buzz", orgNameRetriever(domains[1], domainRoots[1]))
	assert.Equal(t, "", orgNameRetriever(domains[2], domainRoots[0]))

	assert.Equal(t, "", orgNameRetriever("erda.daily.terminus.io", "daily.terminus.io"))
	assert.Equal(t, "", orgNameRetriever("", "daily.terminus.io"))
}
