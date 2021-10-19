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

package accountTable

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-account-manage/accountTable/table"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type comp struct {
	base.DefaultProvider
}

func init() {
	base.InitProviderWithCreator("addon-account-manage", "accountTable",
		func() servicehub.Provider { return &comp{} })
}

func (f *comp) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {

	items := []*AccountItem{
		{
			Username:       "abcd",
			Password:       "******",
			ReferenceCount: 12,
			Creator:        "Mike",
			CreatedAt:      time.Now(),
			IsPrimary:      true,
		},
	}

	var props table.Props
	props.Visible = true
	props.Columns = getTitles()
	props.RowKey = "username"
	c.Props = props

	c.Data = make(map[string]interface{})
	c.Data["list"] = getData(items)

	return nil
}

func getTitles() []*table.ColumnTitle {
	return []*table.ColumnTitle{
		{
			Title:     "账号",
			DataIndex: "username",
		},
		{
			Title:     "密码",
			DataIndex: "password",
		},
		{
			Title:     "被引用的次数",
			DataIndex: "referenceCount",
		},
		{
			Title:     "创建者",
			DataIndex: "creator",
		},
		{
			Title:     "创建时间",
			DataIndex: "createdAt",
		},
		{
			Title:     "操作",
			DataIndex: "operate",
			Width:     180,
		},
	}
}

func getData(items []*AccountItem) []map[string]table.ColumnData {
	var columns []map[string]table.ColumnData
	for _, i := range items {
		columns = append(columns, getDatum(i))
	}
	return columns
}

func getDatum(item *AccountItem) map[string]table.ColumnData {
	datum := make(map[string]table.ColumnData)
	if item.IsPrimary {
		datum["username"] = table.ColumnData{RenderType: "textWithTags", Value: item.Username, Tags: []table.ColumnDataTag{{Tag: "主账号", Color: "green"}}}
	} else {
		datum["username"] = table.ColumnData{RenderType: "text", Value: item.Username}
	}
	datum["password"] = table.ColumnData{RenderType: "text", Value: "******"}
	datum["referenceCount"] = table.ColumnData{RenderType: "linkText", Value: fmt.Sprintf("%d", item.ReferenceCount), Operations: map[string]*table.Operation{
		"click": {
			Key:    "edit",
			Text:   "编辑",
			Reload: false,
			Command: &table.OperationCommand{
				Key: "set",
				State: map[string]string{
					"visible":  "true",
					"username": item.Username,
				},
				Target: "referenceModal",
			},
			ShowIndex:   1,
			Disabled:    false,
			DisabledTip: "无操作权限，请联系项目所有者/项目经理开通操作权限",
		},
	}}
	datum["creator"] = table.ColumnData{RenderType: "text", Value: item.Creator}
	datum["createdAt"] = table.ColumnData{RenderType: "datePicker", Value: item.CreatedAt.Format(time.RFC3339)}
	datum["operate"] = table.ColumnData{RenderType: "tableOperation", Operations: map[string]*table.Operation{
		"setMain": {
			Key:    "setMain",
			Text:   "设置为主账号",
			Reload: true,
			Meta: map[string]string{
				"username": item.Username,
			},
			ShowIndex: 1,
			Confirm:   "主账号将被新引用应用或者已引用应用重启后使用，请确认是否要设置",
		},
		"delete": {
			Key:    "delete",
			Text:   "删除",
			Reload: true,
			Meta: map[string]string{
				"username": item.Username,
			},
			Disabled:    false,
			DisabledTip: "无法删除",
			ShowIndex:   2,
			Confirm:     "是否确认删除",
			SuccessMsg:  "删除成功",
		},
	}}
	return datum
}
