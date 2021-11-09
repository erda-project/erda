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

package addAccountButton

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	addonmysqlpb "github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/common/apis"
)

type comp struct {
	base.DefaultProvider
}

func init() {
	base.InitProviderWithCreator("addon-mysql-account", "addAccountButton",
		func() servicehub.Provider { return &comp{} })
}

func (f *comp) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	userID := apis.GetUserID(ctx)
	ac, err := common.LoadAccountData(ctx)
	if err != nil {
		return err
	}
	props := map[string]interface{}{
		"text":          "一键创建账号",
		"type":          "primary",
		"requestIgnore": []string{"props", "data", "operations"},
		"tipProps": map[string]interface{}{
			"placement": "left",
		},
	}
	if !ac.EditPerm {
		props["disabled"] = true
		props["disabledTip"] = "您没有权限创建账号，请联系项目管理员"
	}
	c.Props = props
	c.Operations = map[string]interface{}{
		"click": cptype.Operation{
			Key:    "addAccount",
			Reload: true,
		},
	}
	addonMySQLSvc := ctx.Value(types.AddonMySQLService).(addonmysqlpb.AddonMySQLServiceServer)
	pg := common.LoadPageDataAccount(ctx)
	switch event.Operation {
	case "addAccount":
		_, err := addonMySQLSvc.GenerateMySQLAccount(utils.NewContextWithHeader(ctx), &addonmysqlpb.GenerateMySQLAccountRequest{
			InstanceId: pg.InstanceID,
			UserID:     userID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
