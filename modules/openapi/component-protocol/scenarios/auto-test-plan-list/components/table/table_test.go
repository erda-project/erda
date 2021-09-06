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

package table

import (
	"testing"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_ConvertSortData(t *testing.T) {
	req := apistructs.TestPlanV2PagingRequest{}
	c := &apistructs.Component{
		State: map[string]interface{}{
			"sorterData": SortData{
				Field: "passRate",
				Order: OrderAscend,
			},
		},
	}
	err := convertSortData(&req, c)
	assert.NoError(t, err)
	want := apistructs.TestPlanV2PagingRequest{
		OrderBy: "pass_rate",
		Asc:     true,
	}
	assert.Equal(t, want, req)

	c = &apistructs.Component{
		State: map[string]interface{}{
			"sorterData": SortData{
				Field: "executeTime",
				Order: OrderDescend,
			},
		},
	}
	err = convertSortData(&req, c)
	assert.NoError(t, err)
	want = apistructs.TestPlanV2PagingRequest{
		OrderBy: "execute_time",
		Asc:     false,
	}
	assert.Equal(t, want, req)
}
