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

package issuestream

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_filterReceiversByOperatorID(t *testing.T) {
	svc := IssueStream{}
	receivers := svc.filterReceiversByOperatorID([]string{"a", "b"}, "b")
	assert.Equal(t, "a", receivers[0])
}

func Test_groupEventContent(t *testing.T) {
	svc := IssueStream{}
	content, err := svc.groupEventContent([]apistructs.IssueStreamType{apistructs.ISTChangeContent}, apistructs.ISTParam{}, "zh")
	assert.NoError(t, err)
	assert.Equal(t, "内容发生变更", content)
}
