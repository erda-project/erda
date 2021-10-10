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

package executeHistoryTable

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/hooks/posthandle"
	"github.com/erda-project/erda/pkg/strutil"
)

type ComponentAction struct {
	base.DefaultProvider

	CodeCoverageSvc *code_coverage.CodeCoverage `json:"-"`
	Type            string                      `json:"type"`
	Data            Data                        `json:"data"`
	Props           map[string]interface{}      `json:"props"`
	State           *State                      `json:"state"`
	Operations      Operations                  `json:"operations"`
}

type Data struct {
	List []ExecuteHistory `json:"list"`
}

type ExecuteHistory struct {
	ID        uint64    `json:"id"`
	Status    Status    `json:"status"`
	Reason    string    `json:"reason"`
	Starter   string    `json:"starter"`
	StartTime string    `json:"startTime"`
	Ender     string    `json:"ender"`
	EndTime   string    `json:"endTime"`
	CoverRate CoverRate `json:"coverRate"`
	Operate   Operate   `json:"operate"`
}

type Status struct {
	RenderType string `json:"renderType"`
	Value      string `json:"value"`
	Status     string `json:"status"`
}

type CoverRate struct {
	RenderType string `json:"renderType"`
	Value      string `json:"value"`
	Tip        string `json:"tip"`
	Status     string `json:"status"`
}

type Operate struct {
	Operations struct {
		Download Download `json:"download"`
	} `json:"operations"`
	RenderType string `json:"renderType"`
}

type Download struct {
	Command  Command `json:"command"`
	Confirm  string  `json:"confirm"`
	Key      string  `json:"key"`
	Reload   bool    `json:"reload"`
	Text     string  `json:"text"`
	Disabled bool    `json:"disabled"`
}

type Command struct {
	JumpOut bool   `json:"jumpOut"`
	Key     string `json:"key"`
	Target  string `json:"target"`
}

type Operations struct {
	ChangePageNo ChangePageNo `json:"changePageNo"`
}

type ChangePageNo struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type State struct {
	PageNo   uint64 `json:"pageNo"`
	PageSize uint64 `json:"pageSize"`
	Total    uint64 `json:"total"`
}

var statusMap = map[string]string{
	"running": "进行中",
	"ready":   "准备中",
	"ending":  "结束中",
	"success": "成功",
	"fail":    "失败",
	"cancel":  "用户取消",
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {

	ca.CodeCoverageSvc = ctx.Value(types.CodeCoverageService).(*code_coverage.CodeCoverage)

	if err := ca.SetState(c); err != nil {
		return err
	}

	if err := ca.setData(ctx, gs); err != nil {
		return err
	}

	ca.SetOperations()
	if err := ca.SetProps(); err != nil {
		return err
	}
	return nil
}

func (ca *ComponentAction) setData(ctx context.Context, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	projectIDStr := sdk.InParams["projectId"].(string)
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		return err
	}
	data, err := ca.CodeCoverageSvc.ListCodeCoverageRecord(apistructs.CodeCoverageListRequest{
		IdentityInfo: apistructs.IdentityInfo{
			UserID: sdk.Identity.GetUserID(),
		},
		ProjectID: projectID,
		PageNo:    ca.State.PageNo,
		PageSize:  ca.State.PageSize,
	})
	if err != nil {
		return err
	}
	ca.State.Total = data.Total
	list := make([]ExecuteHistory, 0)
	userIDs := make([]string, 0)
	for _, v := range data.List {
		disabled := false
		if v.ReportUrl == "" {
			disabled = true
		}
		var timeBegin, timeEnd string
		if v.TimeBegin != nil {
			timeBegin = v.TimeBegin.Format("2006-01-02 15:03:04")
		}
		if v.TimeEnd != nil {
			timeEnd = v.TimeEnd.Format("2006-01-02 15:03:04")
		}
		userIDs = append(userIDs, v.StartExecutor, v.EndExecutor)
		list = append(list, ExecuteHistory{
			ID: v.ID,
			Status: Status{
				RenderType: "textWithBadge",
				Value:      statusMap[v.Status],
				Status:     v.Status,
			},
			Reason:    v.Msg,
			Starter:   v.StartExecutor,
			StartTime: timeBegin,
			Ender:     v.EndExecutor,
			EndTime:   timeEnd,
			CoverRate: CoverRate{
				RenderType: "progress",
				Value:      fmt.Sprintf("%v", v.Coverage),
				Tip:        "",
				Status:     v.Status,
			},
			Operate: Operate{
				Operations: struct {
					Download Download `json:"download"`
				}{Download: Download{
					Command: Command{
						JumpOut: true,
						Key:     "goto",
						Target:  v.ReportUrl,
					},
					Confirm:  "",
					Key:      "download",
					Reload:   false,
					Text:     "下载报告",
					Disabled: disabled,
				}},
				RenderType: "tableOperation",
			},
		})
	}
	userIDs = strutil.DedupSlice(userIDs, true)
	uInfo, err := posthandle.GetUsers(userIDs, false)
	if err != nil {
		return err
	}
	for i := range list {
		list[i].Starter = uInfo[list[i].Starter].Name
		list[i].Ender = uInfo[list[i].Ender].Name
	}
	ca.Data.List = list
	return nil
}

func (ca *ComponentAction) SetState(c *cptype.Component) error {
	b, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state State
	if err = json.Unmarshal(b, &state); err != nil {
		return err
	}
	ca.State = &state
	return nil
}

func (ca *ComponentAction) SetOperations() {
	ca.Operations = Operations{ChangePageNo: ChangePageNo{
		Key:    "changePageNo",
		Reload: true,
	}}
}

func (ca *ComponentAction) SetProps() error {
	props := `
	{
    "pageSizeOptions":[
        "10",
        "20",
        "50",
        "100"
    ],
    "scroll":{
        "x":1000
    },
    "columns":[
        {
            "dataIndex":"status",
            "title":"状态",
            "width":80
        },
        {
            "dataIndex":"coverRate",
            "title":"当前行覆盖率",
            "width":120
        },
        {
            "dataIndex":"reason",
            "title":"原因"
        },
        {
            "dataIndex":"starter",
            "title":"统计开始者",
            "width":100
        },
        {
            "dataIndex":"startTime",
            "title":"开始时间",
            "width":140
        },
        {
            "dataIndex":"ender",
            "title":"统计结束者",
            "width":100
        },
        {
            "dataIndex":"endTime",
            "title":"统计结束时间",
            "width":140
        },
        {
            "dataIndex":"operate",
            "fixed":"right",
            "title":"操作",
            "width":100
        }
    ],
    "rowKey":"id"
}
`
	ca.Props = make(map[string]interface{})
	return json.Unmarshal([]byte(props), &ca.Props)
}

func init() {
	base.InitProviderWithCreator("code-coverage", "executeHistoryTable", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
