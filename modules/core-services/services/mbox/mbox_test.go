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

package mbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
)

func TestSetDuplicateMboxStatus(t *testing.T) {
	tm := time.Now()
	mbox := &model.MBox{
		Title:         "fakeTitle",
		Content:       "fakeContent",
		UserID:        "2",
		Status:        apistructs.MBoxUnReadStatus,
		OrgID:         1,
		ReadAt:        &tm,
		DeduplicateID: "issue-1",
		UnreadCount:   10,
	}
	createReq := &apistructs.CreateMBoxRequest{
		Title:   "fakeTitle2",
		Content: "fakeContent2",
	}
	setDuplicateMbox(mbox, createReq)
	assert.Equal(t, int64(11), mbox.UnreadCount)
	assert.Equal(t, apistructs.MBoxUnReadStatus, mbox.Status)
	assert.Equal(t, "fakeTitle2", mbox.Title)
	assert.Equal(t, "fakeContent2", mbox.Content)
	assert.Equal(t, tm.Format("2006-01-02 15:04:05"), mbox.CreatedAt.Format("2006-01-02 15:04:05"))

	mbox.Status = apistructs.MBoxReadStatus
	setDuplicateMbox(mbox, createReq)
	assert.Equal(t, int64(1), mbox.UnreadCount)
	assert.Equal(t, apistructs.MBoxUnReadStatus, mbox.Status)
	assert.Equal(t, "fakeTitle2", mbox.Title)
	assert.Equal(t, "fakeContent2", mbox.Content)
	assert.Equal(t, tm.Format("2006-01-02 15:04:05"), mbox.CreatedAt.Format("2006-01-02 15:04:05"))
}
