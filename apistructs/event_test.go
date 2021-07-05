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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenEventParams(t *testing.T) {
	issueEventImpl := &IssueEvent{
		EventHeader: EventHeader{
			Event:         "issue",
			Action:        "update",
			OrgID:         "1",
			ProjectID:     "1",
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Content: IssueEventData{
			Title:      "fakeTitle",
			Content:    "hello\\nword\\na",
			AtUserIDs:  "2",
			IssueType:  IssueTypeRequirement,
			StreamType: ISTComment,
			StreamParams: ISTParam{
				CommentTime: "2006-01-02 15:04:05", // comment time
				UserName:    "LiLei",
			},
			Receivers: []string{"2", "3", "4"},
			Params: map[string]string{
				"orgName":     "fakeOrg",
				"projectName": "fakePrj",
				"issueID":     "1",
			},
		},
	}

	params := issueEventImpl.GenEventParams("zh-CN", "https://fake.xx.com")
	assert.Equal(t, params["issueType"], "需求")
	assert.Equal(t, params["issueTitle"], "fakeTitle")
	assert.Equal(t, params["title"], "需求-fakeTitle (fakeOrg/fakePrj 项目)")
	assert.Equal(t, params["projectMboxLink"], "/fakeOrg/dop/projects/1/issues/all")
	assert.Equal(t, params["issueMboxLink"], "/fakeOrg/dop/projects/1/issues/all?id=1&type=REQUIREMENT")
	assert.Equal(t, params["projectEmailLink"], "https://fake.xx.com/fakeOrg/dop/projects/1/issues/all")
	assert.Equal(t, params["issueEmailLink"], "https://fake.xx.com/fakeOrg/dop/projects/1/issues/all?id=1&type=REQUIREMENT")
	assert.Equal(t, params["issueMboxContent"], "LiLei 备注于 2006-01-02 15:04:05\nhello\nword\na")
	assert.Equal(t, params["issueEmailContent"], "LiLei 备注于 2006-01-02 15:04:05</br>hello</br>word</br>a")
	assert.Equal(t, params["mboxDeduplicateID"], "issue-1")
}
