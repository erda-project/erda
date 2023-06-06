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
	assert.Equal(t, uint64(1), model.Common.ID)
	assert.Equal(t, "title", model.Common.IssueTitle)
	assert.Equal(t, "code", model.TaskOnly.TaskType)
	assert.Equal(t, "v-of-cf-1", model.TaskOnly.CustomFields[0].Value)
}

func Test_parseStringSlice(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{
				s: "a，b, c,,",
			},
			want: []string{"a", "b", "c"},
		},
		{
			args: args{
				s: "",
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parseStringSliceByComma(tt.args.s), "parseStringSliceByComma(%v)", tt.args.s)
		})
	}
}

func Test_parseStringLineID(t *testing.T) {
	i, err := parseStringIssueID("a")
	assert.Error(t, err)
	i, err = parseStringIssueID("1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), *i)
	i, err = parseStringIssueID("1.1")
	assert.Error(t, err)
	i, err = parseStringIssueID("")
	assert.NoError(t, err)
	assert.Nil(t, i)
	i, err = parseStringIssueID("L254")
	assert.NoError(t, err)
	assert.Equal(t, int64(-254), *i)
	i, err = parseStringIssueID("L")
	assert.Error(t, err)
	i, err = parseStringIssueID("L-1")
	assert.Error(t, err)
	i, err = parseStringIssueID("-1")
	assert.Error(t, err)
}
