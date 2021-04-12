// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package appsitemanage

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EdgeAppDetailItem struct {
	ID           int                           `json:"id"`
	SiteName     map[string]interface{}        `json:"siteName"`
	DeployStatus apistructs.EdgeTextBadge      `json:"deployStatus"`
	Operate      apistructs.EdgeItemOperations `json:"operate"`
}

type EdgeAppDetailExItem struct {
	ID        int64                         `json:"id"`
	IP        string                        `json:"ip"`
	Address   string                        `json:"address"`
	Status    apistructs.EdgeTextBadge      `json:"status"`
	CreatedAt string                        `json:"createdAt"`
	Operate   apistructs.EdgeItemOperations `json:"operate"`
}

func (c *ComponentList) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	if c.component.State == nil {
		c.component.State = make(map[string]interface{})
	}

	identity := c.ctxBundle.Identity

	if strings.HasPrefix(event.Operation.String(), apistructs.EdgeOperationChangePage) ||
		event.Operation == apistructs.InitializeOperation ||
		event.Operation == apistructs.RenderingOperation {
		if err := c.OperateChangePage(false, "", identity); err != nil {
			return fmt.Errorf("change page operation error: %v", err)
		}
	} else if event.Operation == apistructs.EdgeOperationOffline {
		if err := c.OperateOffline(event.OperationData, identity); err != nil {
			return fmt.Errorf("offline operation error: %v", err)
		}
	} else if event.Operation == apistructs.EdgeOperationRestart {
		if err := c.OperateRestart(event.OperationData, identity); err != nil {
			return fmt.Errorf("restart operation error: %v", err)
		}
	}

	c.component.Operations = getOperations()
	c.component.Props = getProps()

	return nil
}

func getProps() apistructs.EdgeTableProps {
	return apistructs.EdgeTableProps{
		PageSizeOptions: []string{"10", "20", "50", "100"},
		RowKey:          "siteName",
		Columns: []apistructs.EdgeColumns{
			{Title: "站点名称", DataIndex: "siteName", Width: 150},
			{Title: "部署状态", DataIndex: "deployStatus", Width: 150},
			{Title: "操作", DataIndex: "operate", Width: 150},
		},
	}
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"changePageNo": apistructs.EdgeOperation{
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": apistructs.EdgeOperation{
			Key:    "changePageSize",
			Reload: true,
		},
	}
}

func getSiteItemOperate(inParam EdgeAppSiteInParam, siteName string, isAllOperate bool) apistructs.EdgeItemOperations {
	return apistructs.EdgeItemOperations{
		RenderType: "tableOperation",
		Operations: map[string]apistructs.EdgeItemOperation{
			apistructs.EdgeOperationRestart: {
				ShowIndex:   1,
				Key:         apistructs.EdgeOperationRestart,
				Text:        "重启",
				Confirm:     "是否确认重启站点",
				Disabled:    !isAllOperate,
				DisabledTip: "应用未部署成功, 无法重启",
				Reload:      true,
				Meta: map[string]interface{}{
					"appID":    inParam.ID,
					"siteName": siteName,
				},
			},
			apistructs.EdgeOperationOffline: {
				ShowIndex:   2,
				Key:         apistructs.EdgeOperationOffline,
				Text:        "下线",
				Confirm:     "是否确认下线站点",
				Disabled:    !isAllOperate,
				DisabledTip: "应用未部署成功, 无法下线",
				Reload:      true,
				Meta: map[string]interface{}{
					"appID":    inParam.ID,
					"siteName": siteName,
				},
			},
		},
	}
}

func renderSiteName(id int64, siteName, appName string) map[string]interface{} {
	return map[string]interface{}{
		"renderType": "linkText",
		"value":      siteName,
		"operations": apistructs.EdgeOperations{
			"click": apistructs.EdgeOperation{
				Key:    "gotoSiteManage",
				Reload: false,
				Command: apistructs.EdgeJumpCommand{
					Key:     "goto",
					Target:  "edgeAppSiteIpManage",
					JumpOut: false,
					State: apistructs.EdgeJumpCommandState{
						Params: map[string]interface{}{
							"id":       id,
							"siteName": siteName,
							"appName":  appName,
						},
						Query: map[string]interface{}{
							"appName": appName,
						},
						Visible: false,
					},
				},
			},
		},
	}
}
