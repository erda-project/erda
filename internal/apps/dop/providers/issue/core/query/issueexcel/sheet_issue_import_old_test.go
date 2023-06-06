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

package issueexcel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

func Test_tryToMatchCustomFieldNameToIssueType(t *testing.T) {
	cfNames := []string{"测试", "测试人员", "提测时间", "测试人员", "测试自定义字段关联", "测试", "custom2"}
	customFieldMap := map[pb.PropertyIssueTypeEnum_PropertyIssueType][]*pb.IssuePropertyIndex{
		pb.PropertyIssueTypeEnum_REQUIREMENT: {
			{
				PropertyIssueType: pb.PropertyIssueTypeEnum_REQUIREMENT,
				PropertyName:      "提测时间",
			},
			{
				PropertyIssueType: pb.PropertyIssueTypeEnum_REQUIREMENT,
				PropertyName:      "测试人员",
			},
			{
				PropertyIssueType: pb.PropertyIssueTypeEnum_REQUIREMENT,
				PropertyName:      "测试自定义字段关联",
			},
		},
		pb.PropertyIssueTypeEnum_TASK: {
			{
				PropertyIssueType: pb.PropertyIssueTypeEnum_TASK,
				PropertyName:      "测试",
			},
			{
				PropertyIssueType: pb.PropertyIssueTypeEnum_TASK,
				PropertyName:      "custom2",
			},
		},
		pb.PropertyIssueTypeEnum_BUG: {
			{
				PropertyIssueType: pb.PropertyIssueTypeEnum_BUG,
				PropertyName:      "测试",
			},
			{
				PropertyIssueType: pb.PropertyIssueTypeEnum_BUG,
				PropertyName:      "测试人员",
			},
		},
	}
	result, err := tryToMatchCustomFieldNameToIssueType(cfNames, customFieldMap)
	assert.NoError(t, err)
	assert.True(t, result[0] == pb.PropertyIssueTypeEnum_BUG)
	assert.True(t, result[1] == pb.PropertyIssueTypeEnum_BUG)
	assert.True(t, result[2] == pb.PropertyIssueTypeEnum_REQUIREMENT)
	assert.True(t, result[3] == pb.PropertyIssueTypeEnum_REQUIREMENT)
	assert.True(t, result[4] == pb.PropertyIssueTypeEnum_REQUIREMENT)
	assert.True(t, result[5] == pb.PropertyIssueTypeEnum_TASK)
	assert.True(t, result[6] == pb.PropertyIssueTypeEnum_TASK)

	cfNames = []string{"测试", "测试人员"}
	result, err = tryToMatchCustomFieldNameToIssueType(cfNames, customFieldMap)
	assert.NoError(t, err)
	assert.True(t, result[0] == pb.PropertyIssueTypeEnum_BUG)
	assert.True(t, result[1] == pb.PropertyIssueTypeEnum_BUG)

	cfNames = []string{"测试", "custom2"}
	result, err = tryToMatchCustomFieldNameToIssueType(cfNames, customFieldMap)
	assert.NoError(t, err)
	assert.True(t, result[0] == pb.PropertyIssueTypeEnum_TASK)
	assert.True(t, result[1] == pb.PropertyIssueTypeEnum_TASK)

	cfNames = []string{"测试", "custom3"}
	result, err = tryToMatchCustomFieldNameToIssueType(cfNames, customFieldMap)
	assert.Error(t, err)

	cfNames = []string{"测试"}
	result, err = tryToMatchCustomFieldNameToIssueType(cfNames, customFieldMap)
	assert.Error(t, err)
}
