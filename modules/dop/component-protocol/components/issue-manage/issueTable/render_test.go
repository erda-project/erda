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

package issueTable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
)

func TestGetTotalPage(t *testing.T) {
	data := []struct {
		total    uint64
		pageSize uint64
		page     uint64
	}{
		{10, 0, 0},
		{0, 10, 0},
		{20, 10, 2},
		{21, 10, 3},
	}
	for _, v := range data {
		assert.Equal(t, getTotalPage(v.total, v.pageSize), v.page)
	}
}

func Test_resetPageInfo(t *testing.T) {
	type args struct {
		req   *apistructs.IssuePagingRequest
		state map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "reset",
			args: args{
				req: &apistructs.IssuePagingRequest{},
				state: map[string]interface{}{
					"pageNo":   float64(1),
					"pageSize": float64(2),
				},
			},
		},
	}

	expected := []*apistructs.IssuePagingRequest{
		{
			PageNo:   1,
			PageSize: 2,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetPageInfo(tt.args.req, tt.args.state)
			assert.Equal(t, expected[i], tt.args.req)
		})
	}
}

func Test_getPrefixIcon(t *testing.T) {
	assert.Equal(t, "ISSUE_ICON.issue.abc", getPrefixIcon("abc"))
	assert.Equal(t, "ISSUE_ICON.issue.123", getPrefixIcon("123"))
	assert.Equal(t, "ISSUE_ICON.issue.", getPrefixIcon(""))
}

func Test_resetPageNoByFilterCondition(t *testing.T) {
	assert.False(t, resetPageNoByFilterCondition("a", struct {
		a string
	}{}, map[string]interface{}{"a": "111"}))
	assert.True(t, resetPageNoByFilterCondition("b", struct {
		a string
	}{}, map[string]interface{}{"a": "111"}))
}

type MockTran struct {
	i18n.Translator
}

func (m *MockTran) Text(lang i18n.LanguageCodes, key string) string {
	return ""
}

func (m *MockTran) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return ""
}

func Test_buildTableItem(t *testing.T) {
	ctx := context.WithValue(context.Background(), cptype.GlobalInnerKeyCtxSDK, &cptype.SDK{Tran: &MockTran{}})
	ca := ComponentAction{}
	i := ca.buildTableItem(ctx, &apistructs.Issue{})
	assert.NotNil(t, i)
}
