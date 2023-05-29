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

	"github.com/erda-project/erda/pkg/excel"
)

func Test_decodeMapToIssueSheetModel(t *testing.T) {
	autoCompleteUUID := func(parts ...string) IssueSheetColumnUUID {
		uuid := NewIssueSheetColumnUUID(parts...)
		return IssueSheetColumnUUID(uuid.String())
	}
	data := DataForFulfill{}
	models, err := data.decodeMapToIssueSheetModel(map[IssueSheetColumnUUID]excel.Column{
		autoCompleteUUID("Common", "ID"): {
			{Value: "1"},
		},
		autoCompleteUUID("Common", "IssueTitle"): {
			{Value: "title"},
		},
		autoCompleteUUID("TaskOnly", "TaskType"): {
			{Value: "code"},
		},
		autoCompleteUUID("TaskOnly", "CustomFields", "cf-1"): {
			{Value: "v-of-cf-1"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(models))
	model := models[0]
	assert.Equal(t, int64(1), model.Common.ID)
	assert.Equal(t, "title", model.Common.IssueTitle)
	assert.Equal(t, "code", model.TaskOnly.TaskType)
	assert.Equal(t, "v-of-cf-1", model.TaskOnly.CustomFields[0].Value)
}
