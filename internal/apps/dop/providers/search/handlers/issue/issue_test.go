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

package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

func Test_convertIssueToValue(t *testing.T) {
	issue := &pb.Issue{
		Id:      1,
		Title:   "issue-1",
		Content: "issue-1-content",
	}
	handler := &issueSearch{}
	value, err := handler.convertIssueToValue(issue)
	assert.NoError(t, err)
	dat := value.GetStructValue()
	assert.Equal(t, "issue-1", dat.GetFields()["title"].GetStringValue())
	assert.Equal(t, "", dat.GetFields()["content"].GetStringValue())
}

func Test_genIssueLink(t *testing.T) {
	handler := &issueSearch{}
	link := handler.genIssueLink("erda", &pb.Issue{
		Id:        1,
		ProjectID: 2,
		Type:      pb.IssueTypeEnum_BUG,
	})
	assert.Equal(t, "/erda/dop/projects/2/issues/all?id=1&type=BUG", link)
}
