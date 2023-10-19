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

package sheet_issue

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
)

func Test_decodeMapToIssueSheetModel(t *testing.T) {
	autoCompleteUUID := func(parts ...string) IssueSheetColumnUUID {
		uuid := NewIssueSheetColumnUUID(parts...)
		return IssueSheetColumnUUID(uuid.String())
	}
	data := &vars.DataForFulfill{
		StageMap: map[query.IssueStage]string{
			query.IssueStage{
				Type:  pb.IssueTypeEnum_TASK.String(),
				Value: "code",
			}: "code",
		},
	}
	models, err := decodeMapToIssueSheetModel(data, map[IssueSheetColumnUUID]excel.Column{
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
			assert.Equalf(t, tt.want, vars.ParseStringSliceByComma(tt.args.s), "parseStringSliceByComma(%v)", tt.args.s)
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

func Test_removeEmptySheetRows(t *testing.T) {
	rows := [][]string{
		{"", "", ""},
		{"a", "b", ""},
		{"", "", ""},
		{"c", "", ""},
	}
	removeEmptySheetRows(&rows)
	assert.True(t, len(rows) == 2)
	assert.Equal(t, []string{"a", "b", ""}, rows[0])
	assert.Equal(t, []string{"c", "", ""}, rows[1])
}

func Test_parseStringTime(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *time.Time
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "normal",
			args: args{
				s: "2021-07-12 12:00:00",
			},
			want: func() *time.Time {
				t := time.Date(2021, 7, 12, 12, 0, 0, 0, time.Local)
				return &t
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "normal slash",
			args: args{
				s: "2021/07/12 12:00:00",
			},
			want: func() *time.Time {
				t := time.Date(2021, 7, 12, 12, 0, 0, 0, time.Local)
				return &t
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "short slash",
			args: args{
				s: "2021/7/12 0:00:00",
			},
			want: func() *time.Time {
				t := time.Date(2021, 7, 12, 0, 0, 0, 0, time.Local)
				return &t
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "empty",
			args: args{
				s: "",
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "-",
			args: args{
				s: `2021\-07\-12 12:00:00`,
			},
			want: func() *time.Time {
				t := time.Date(2021, 7, 12, 12, 0, 0, 0, time.Local)
				return &t
			}(),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStringTime(tt.args.s)
			if !tt.wantErr(t, err, fmt.Sprintf("parseStringTime(%v)", tt.args.s)) {
				return
			}
			assert.Equalf(t, tt.want, got, "parseStringTime(%v)", tt.args.s)
		})
	}
}

func Test_autoFillEmptyRowCells(t *testing.T) {
	rows := [][]string{
		{"a", "b", "c"},
		{"a"},
		{"1", "2"},
	}
	autoFillEmptyRowCells(&rows, 3)
	assert.Equal(t, 3, len(rows[0]))

	assert.Equal(t, 3, len(rows[1]))
	assert.Equal(t, "", rows[1][1])
	assert.Equal(t, "", rows[1][2])

	assert.Equal(t, 3, len(rows[2]))
	assert.Equal(t, "", rows[2][2])
}
