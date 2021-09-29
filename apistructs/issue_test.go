// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apistructs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetRelatedIssueIDs(t *testing.T) {
	iss := Issue{}
	iss.SetRelatedIssueIDs("1001,1002")
	relatedIDs := iss.GetRelatedIssueIDs()
	assert.Equal(t, 2, len(relatedIDs))
	assert.Equal(t, uint64(1001), relatedIDs[0])
	assert.Equal(t, uint64(1002), relatedIDs[1])
}
