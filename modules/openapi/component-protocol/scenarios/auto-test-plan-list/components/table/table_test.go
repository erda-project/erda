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
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"
	"golang.org/x/net/context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
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

func Test_executeTime(t *testing.T) {
	nowTime := time.Now()
	var data = []*apistructs.TestPlanV2{
		{
			ExecuteTime: nil,
		},
		{
			ExecuteTime: &nowTime,
		},
	}
	want := []string{
		"",
		nowTime.Format("2006-01-02 15:04:05"),
	}
	for i := range data {
		executeTime := convertExecuteTime(data[i])
		assert.Equal(t, executeTime, want[i])
	}
}

func Test_Render(t *testing.T) {
	bdl := &bundle.Bundle{}
	cb := protocol.ContextBundle{
		Bdl: bdl,
		InParams: map[string]interface{}{
			"projectId": 1,
		},
	}
	ctx := context.WithValue(context.Background(), protocol.GlobalInnerKeyCtxBundle.String(), cb)
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "PagingTestPlansV2",
		func(b *bundle.Bundle, req apistructs.TestPlanV2PagingRequest) (*apistructs.TestPlanV2PagingResponseData, error) {
			list := []*apistructs.TestPlanV2{
				{
					PassRate: 10,
				},
				{
					PassRate: 0,
				},
			}

			return &apistructs.TestPlanV2PagingResponseData{
				List: list,
			}, nil
		})
	defer monkey.UnpatchAll()
	p := &TestPlanManageTable{}
	c := &apistructs.Component{}
	gs := &apistructs.GlobalStateData{}
	err := p.Render(ctx, c, apistructs.ComponentProtocolScenario{},
		apistructs.ComponentEvent{
			Operation:     "ooo",
			OperationData: nil,
		}, gs)
	assert.NoError(t, err)
	list := c.Data["list"].([]TableItem)
	want := []string{"10", "0"}
	for i := range list {
		assert.Equal(t, list[i].PassRate.Value, want[i])
	}

}
