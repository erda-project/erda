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

package applicationlist

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EdgeAPPItem struct {
	ID              int64                         `json:"id"`
	ApplicationName apistructs.EdgeItemOperations `json:"appName"`
	Cluster         string                        `json:"cluster"`
	DeployResource  string                        `json:"deployResource"`
	Operate         apistructs.EdgeItemOperations `json:"operate"`
}

func (c ComponentList) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var (
		isKeyValueListVisible bool
	)

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	orgID, err := strconv.ParseInt(c.ctxBundle.Identity.OrgID, 10, 64)
	if err != nil {
		return fmt.Errorf("component %s parse org id error: %v", component.Name, err)
	}

	if c.component.State == nil {
		c.component.State = map[string]interface{}{}
	}

	c.component.State["addAppDrawerVisible"] = false
	c.component.State["appConfigFormVisible"] = false
	c.component.State["formClear"] = false
	c.component.State["operationType"] = ""
	c.component.State["appID"] = 0

	identity := c.ctxBundle.Identity

	if strings.HasPrefix(event.Operation.String(), apistructs.EdgeOperationChangePage) ||
		event.Operation == apistructs.InitializeOperation ||
		event.Operation == apistructs.RenderingOperation {
		err = c.OperateChangePage(orgID, identity)
		if err != nil {
			return err
		}
	} else if event.Operation.String() == apistructs.EdgeOperationDelete {
		err = c.OperateDelete(orgID, event.OperationData, identity)
		if err != nil {
			return err
		}
	} else if event.Operation.String() == apistructs.EdgeOperationUpdate {
		err = c.OperateReload(event.OperationData)
		if err != nil {
			return fmt.Errorf("operation reload error: %v", err)
		}
		c.component.State["operationType"] = apistructs.EdgeOperationUpdate
	} else if event.Operation.String() == apistructs.EdgeOperationViewDetail {
		err = c.OperateReload(event.OperationData)
		if err != nil {
			return fmt.Errorf("operation reload error: %v", err)
		}
		c.component.State["operationType"] = apistructs.EdgeOperationViewDetail
		isKeyValueListVisible = true
	}

	c.component.State["keyValueListVisible"] = isKeyValueListVisible
	c.component.State["keyValueListTitleVisible"] = isKeyValueListVisible
	c.component.Props = getProps()
	c.component.Operations = getOperations()

	return nil
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

func getProps() apistructs.EdgeTableProps {
	return apistructs.EdgeTableProps{
		PageSizeOptions: []string{"10", "20", "50", "100"},
		RowKey:          "id",
		Columns: []apistructs.EdgeColumns{
			{Title: "应用名称", DataIndex: "appName", Width: 150},
			{Title: "所属集群", DataIndex: "cluster", Width: 150},
			{Title: "部署来源", DataIndex: "deployResource", Width: 150},
			{Title: "操作", DataIndex: "operate", Width: 150},
		},
	}
}

func getAPPItemOperate(appName, deployResource string, id int64) apistructs.EdgeItemOperations {
	return apistructs.EdgeItemOperations{
		RenderType: "tableOperation",
		Operations: map[string]apistructs.EdgeItemOperation{
			apistructs.EdgeOperationViewDetail: {
				ShowIndex: 1,
				Key:       apistructs.EdgeOperationViewDetail,
				Text:      "详情",
				Reload:    true,
				Meta:      map[string]interface{}{"id": id},
				Command: apistructs.EdgeJumpCommand{
					Key:    "set",
					Target: "addAppDrawer",
					State: apistructs.EdgeJumpCommandState{
						FormData: map[string]interface{}{
							"id":             id,
							"appName":        appName,
							"deployResource": deployResource,
						},
					},
				},
			},
			apistructs.EdgeOperationUpdate: {
				ShowIndex: 2,
				Key:       apistructs.EdgeOperationUpdate,
				Text:      "编辑",
				Reload:    true,
				Meta:      map[string]interface{}{"id": id},
				Command: apistructs.EdgeJumpCommand{
					Key:    "set",
					Target: "addAppDrawer",
					State: apistructs.EdgeJumpCommandState{
						FormData: map[string]interface{}{
							"id":             id,
							"appName":        appName,
							"deployResource": deployResource,
						},
					},
				},
			},
			apistructs.EdgeOperationDelete: {
				ShowIndex:   3,
				Key:         apistructs.EdgeOperationDelete,
				Text:        "删除",
				Confirm:     "是否确认删除",
				Disabled:    false,
				DisabledTip: "无法删除",
				Reload:      true,
				Meta:        map[string]interface{}{"id": id},
			},
		},
	}
}

func renderAppName(name string, id int64) apistructs.EdgeItemOperations {
	return apistructs.EdgeItemOperations{
		RenderType: "linkText",
		Value:      name,
		Operations: map[string]apistructs.EdgeItemOperation{
			"click": {
				Key:    "gotoDetail",
				Reload: false,
				Command: apistructs.EdgeJumpCommand{
					Key:     "goto",
					Target:  "ecpAppSiteManage",
					JumpOut: false,
					State: apistructs.EdgeJumpCommandState{
						Params: map[string]interface{}{
							"id": id,
						},
						Query: map[string]interface{}{
							"appName": name,
						},
					},
				},
			},
		},
	}
}
