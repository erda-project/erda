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

package apistructs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetRelatingIssueIDs(t *testing.T) {
	iss := Issue{}
	iss.SetRelatingIssueIDs("1301,1302")
	relatingIDs := iss.GetRelatingIssueIDs()
	assert.Equal(t, 2, len(relatingIDs))
	assert.Equal(t, uint64(1301), relatingIDs[0])
	assert.Equal(t, uint64(1302), relatingIDs[1])
}

func TestSetRelatedIssueIDs(t *testing.T) {
	iss := Issue{}
	iss.SetRelatedIssueIDs("1001,1002")
	relatedIDs := iss.GetRelatedIssueIDs()
	assert.Equal(t, 2, len(relatedIDs))
	assert.Equal(t, uint64(1001), relatedIDs[0])
	assert.Equal(t, uint64(1002), relatedIDs[1])
}
