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

package viewPassword

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	addonmysqlpb "github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/accountTable/table"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type comp struct {
	base.DefaultProvider

	ac *common.AccountData
}

func init() {
	base.InitProviderWithCreator("addon-mysql-account", "viewPassword",
		func() servicehub.Provider { return &comp{} })
}

func (f *comp) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	pg := common.LoadPageDataAccount(ctx)
	if !pg.ShowViewPasswordModal {
		state := make(map[string]interface{})
		state["visible"] = false
		c.State = state
		c.Props = nil
		c.Data = nil
		return nil
	}

	ac, err := common.LoadAccountData(ctx)
	if err != nil {
		return err
	}
	f.ac = ac

	var props table.Props
	props.Bordered = table.False()
	props.Columns = getTitles()
	props.RowKey = "label"
	c.Props = props

	c.Data = make(map[string]interface{})
	c.Data["list"] = getData(ac.AccountMap[pg.AccountID])

	state := make(map[string]interface{})
	state["visible"] = true
	c.State = state

	return nil
}

func getTitles() []*table.ColumnTitle {
	return []*table.ColumnTitle{
		{
			DataIndex: "label",
			Width:     80,
		},
		{
			DataIndex: "value",
		},
	}
}

func getData(account *addonmysqlpb.MySQLAccount) []map[string]interface{} {
	if account == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"label": "账号",
			"value": account.Username,
		},
		{
			"label": "密码",
			"value": account.Password,
		},
	}
}
