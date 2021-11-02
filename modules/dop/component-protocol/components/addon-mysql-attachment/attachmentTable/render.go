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

package attachmentTable

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	addonmysqlpb "github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/accountTable/table"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type comp struct {
	base.DefaultProvider

	pg *common.PageDataAttachment
	ac *common.AccountData
}

func init() {
	base.InitProviderWithCreator("addon-mysql-consumer", "attachmentTable",
		func() servicehub.Provider { return &comp{} })
}

func (f *comp) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	f.pg = common.LoadPageDataAttachment(ctx)

	ac, err := common.LoadAccountData(ctx)
	if err != nil {
		return err
	}
	f.ac = ac

	switch event.Operation {
	case "showConfig":
		f.pg.ShowConfigPanel = true
		f.pg.AttachmentID = event.OperationData["meta"].(map[string]interface{})["id"].(string)
	case "editAttachment":
		f.pg.ShowEditFormModal = true
		f.pg.AttachmentID = event.OperationData["meta"].(map[string]interface{})["id"].(string)
	}

	var props table.Props
	props.Columns = getTitles()
	props.RowKey = "id"
	c.Props = props

	c.Data = make(map[string]interface{})
	c.Data["list"] = f.getData()

	return nil
}

func getTitles() []*table.ColumnTitle {
	return []*table.ColumnTitle{
		{
			Title:     "应用",
			DataIndex: "appName",
		},
		{
			Title:     "实例",
			DataIndex: "runtime",
		},
		{
			Title:     "账号",
			DataIndex: "account",
		},
		{
			Title:     "服务参数",
			DataIndex: "config",
		},
		{
			Title:     "应用部署详情",
			DataIndex: "detail",
		},
		{
			Title:     "操作",
			DataIndex: "operate",
			Width:     180,
		},
	}
}

func (f *comp) getData() []map[string]table.ColumnData {
	var columns []map[string]table.ColumnData
	for _, i := range f.ac.Attachments {
		if !f.getFilter().Match(i) {
			continue
		}
		columns = append(columns, f.getDatum(i))
	}
	return columns
}

func (f *comp) getFilter() table.Matcher {
	return table.And(
		table.ToMatcher(f.accountFilter),
		table.ToMatcher(f.stateFilter),
		table.ToMatcher(f.appFilter),
	)
}

func (f *comp) accountFilter(v interface{}) bool {
	opts := f.pg.FilterValues.StringSlice("account")
	if opts == nil || len(opts) == 0 {
		return true
	}
	item := v.(*addonmysqlpb.Attachment)
	return strutil.Exist(opts, item.AccountId) || (item.AccountState == "PRE" && strutil.Exist(opts, item.PreAccountId))
}

func (f *comp) stateFilter(v interface{}) bool {
	opts := f.pg.FilterValues.StringSlice("state")
	if opts == nil || len(opts) == 0 {
		return true
	}
	item := v.(*addonmysqlpb.Attachment)
	for _, opt := range opts {
		if opt == "PRE" && item.AccountState == "PRE" {
			return true
		}
		if opt == "CUR" && (item.AccountState == "CUR" || item.AccountState == "") {
			return true
		}
	}
	return false
}

func (f *comp) appFilter(v interface{}) bool {
	opts := f.pg.FilterValues.StringSlice("app")
	if opts == nil || len(opts) == 0 {
		return true
	}
	item := v.(*addonmysqlpb.Attachment)
	return strutil.Exist(opts, item.AppId)
}

func (f *comp) getDatum(item *addonmysqlpb.Attachment) map[string]table.ColumnData {
	switching := item.AccountState == "PRE"

	datum := make(map[string]table.ColumnData)
	datum["appName"] = table.ColumnData{RenderType: "text", Value: f.ac.GetAppName(item.AppId)}

	var accountText string
	if switching {
		accountText = fmt.Sprintf("%s -> %s",
			f.ac.GetAccountName(item.PreAccountId), f.ac.GetAccountName(item.AccountId))
	} else {
		accountText = f.ac.GetAccountName(item.AccountId)
	}
	datum["account"] = table.ColumnData{RenderType: "text", Value: accountText}

	if switching {
		datum["runtime"] = table.ColumnData{RenderType: "textWithTags", Value: item.RuntimeName, Tags: []table.ColumnDataTag{
			{
				Tag:   "账号切换中",
				Color: "orange",
			},
		}}
	} else {
		datum["runtime"] = table.ColumnData{RenderType: "text", Value: item.RuntimeName}
	}

	datum["config"] = table.ColumnData{RenderType: "tableOperation", Operations: map[string]*table.Operation{
		"showConfig": {
			Key:    "showConfig",
			Text:   "查看详情",
			Reload: true,
			Meta: map[string]string{
				"id": fmt.Sprintf("%d", item.Id),
			},
		},
	}}
	datum["detail"] = table.ColumnData{RenderType: "linkText", Value: "部署详情", Operations: map[string]*table.Operation{
		"click": {
			Key:    "gotoRuntimeDetail",
			Text:   "查看详情",
			Reload: false,
			Command: &table.OperationCommand{
				Key: "goto",
				State: map[string]interface{}{
					"params": map[string]interface{}{
						"projectId": "TODO",
						"appId":     "",
						"runtimeId": "",
					},
				},
				Target: "runtimeDetailRoot",
			},
		},
	}}
	datum["operate"] = table.ColumnData{RenderType: "tableOperation", Operations: map[string]*table.Operation{
		"edit": {
			Key:    "editAttachment",
			Text:   "编辑",
			Reload: true,
			Meta: map[string]string{
				"id": fmt.Sprintf("%d", item.Id),
			},
		},
	}}
	return datum
}
