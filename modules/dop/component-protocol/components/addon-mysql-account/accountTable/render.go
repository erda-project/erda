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
	"encoding/base64"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	addonmysqlpb "github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/accountTable/table"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/strutil"
)

type comp struct {
	base.DefaultProvider

	ac      *common.AccountData
	pg      *common.PageDataAccount
	userIDs []string
}

func init() {
	base.InitProviderWithCreator("addon-mysql-account", "accountTable",
		func() servicehub.Provider { return &comp{} })
}

func (f *comp) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	f.pg = common.LoadPageDataAccount(ctx)

	ac, err := common.LoadAccountData(ctx)
	if err != nil {
		return err
	}
	f.ac = ac

	switch event.Operation {
	case "viewPassword":
		f.pg.ShowViewPasswordModal = true
		f.pg.AccountID = event.OperationData["meta"].(map[string]interface{})["id"].(string)
	case "delete":
		//f.pg.ShowDeleteModal = true
		accountID := event.OperationData["meta"].(map[string]interface{})["id"].(string)
		addonMySQLSvc := ctx.Value(types.AddonMySQLService).(addonmysqlpb.AddonMySQLServiceServer)
		_, err := addonMySQLSvc.DeleteMySQLAccount(ctx, &addonmysqlpb.DeleteMySQLAccountRequest{
			InstanceId: f.pg.InstanceID,
			Id:         accountID,
		})
		if err != nil {
			return err
		}

		// do reload data
		f.pg, err = common.InitPageDataAccount(ctx)
		if err != nil {
			return err
		}

		f.ac, err = common.InitAccountData(ctx, f.pg.InstanceID, f.pg.ProjectID)
		if err != nil {
			return err
		}
	}

	var props table.Props
	props.Columns = getTitles()
	props.RowKey = "id"
	c.Props = props

	c.Data = make(map[string]interface{})
	c.Data["list"] = f.getData()

	(*gs)[string(cptype.GlobalInnerKeyUserIDs)] = strutil.DedupSlice(f.userIDs, true)

	return nil
}

func getTitles() []*table.ColumnTitle {
	return []*table.ColumnTitle{
		{
			Title:     "账号",
			DataIndex: "username",
		},
		{
			Title:     "使用状态",
			DataIndex: "attachments",
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

func (f *comp) getData() []map[string]table.ColumnData {
	var columns []map[string]table.ColumnData
	for _, i := range f.ac.Accounts {
		if !f.getFilter().Match(i) {
			continue
		}
		datum := f.getDatum(i)
		if datum == nil {
			continue
		}
		columns = append(columns, datum)
	}
	return columns
}

func (f *comp) creatorFilter(v interface{}) bool {
	opts := f.pg.FilterValues.StringSlice("creator")
	if opts == nil || len(opts) == 0 {
		return true
	}
	return strutil.Exist(opts, v.(*addonmysqlpb.MySQLAccount).Creator)
}

func (f *comp) statusFilter(v interface{}) bool {
	item := v.(*addonmysqlpb.MySQLAccount)
	opts := f.pg.FilterValues.StringSlice("status")
	if opts == nil || len(opts) == 0 {
		return true
	}
	for _, status := range opts {
		cnt := f.ac.AccountRefCount[item.Id]
		if status == "YES" && cnt > 0 {
			return true
		}
		if status == "NO" && cnt == 0 {
			return true
		}
	}
	return false
}

func (f *comp) getFilter() table.Matcher {
	return table.And(
		table.ToMatcher(f.creatorFilter),
		table.ToMatcher(f.statusFilter),
	)
}

func (f *comp) getDatum(item *addonmysqlpb.MySQLAccount) map[string]table.ColumnData {
	datum := make(map[string]table.ColumnData)
	datum["username"] = table.ColumnData{RenderType: "text", Value: item.Username}

	cnt := f.ac.AccountRefCount[item.Id]
	datum["attachments"] = table.ColumnData{RenderType: "linkText", Value: fmt.Sprintf("使用中 (%d)", cnt), Operations: map[string]*table.Operation{
		"click": {
			Key:    "gotoMysqlUserManager",
			Reload: false,
			Command: &table.OperationCommand{
				Key: "goto",
				State: map[string]interface{}{
					"params": map[string]interface{}{
						"projectId":  fmt.Sprintf("%d", f.pg.ProjectID),
						"instanceId": f.pg.InstanceID,
						"filter__urlQuery": base64.StdEncoding.EncodeToString(
							[]byte(fmt.Sprintf(`{"account":["%s"]}`, item.Id))),
					},
				},
				Target: "addonPlatformMysqlConsumer",
			},
		},
	}}
	if cnt == 0 {
		datum["attachments"] = table.ColumnData{RenderType: "text", Value: "未被使用"}
	}

	datum["creator"] = table.ColumnData{RenderType: "userAvatar", Value: item.Creator}
	f.userIDs = append(f.userIDs, item.Creator)
	datum["createdAt"] = table.ColumnData{RenderType: "datePicker", Value: item.CreateAt.AsTime().Format(time.RFC3339)}
	datum["operate"] = table.ColumnData{RenderType: "tableOperation", Operations: map[string]*table.Operation{
		"viewPassword": {
			Key:    "viewPassword",
			Text:   "查看密码",
			Reload: true,
			Meta: map[string]string{
				"id": item.Id,
			},
			ShowIndex: 1,
		},
		"delete": {
			Key:    "delete",
			Text:   "删除",
			Reload: true,
			Meta: map[string]string{
				"id": item.Id,
			},
			Disabled:    cnt > 0,
			DisabledTip: "无法删除",
			ShowIndex:   2,
			Confirm:     "是否确认删除",
			SuccessMsg:  "删除成功",
		},
	}}
	return datum
}
