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
	"testing"

	"github.com/stretchr/testify/assert"

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
