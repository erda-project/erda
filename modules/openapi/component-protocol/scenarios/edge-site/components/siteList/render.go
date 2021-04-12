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

package sitelist

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EdgeSiteItem struct {
	ID             int64                         `json:"id"`
	SiteName       apistructs.EdgeItemOperations `json:"siteName"`
	NodeNum        string                        `json:"nodeNum"`
	RelatedCluster string                        `json:"relatedCluster"`
	Operate        apistructs.EdgeItemOperations `json:"operate"`
}

func (c *ComponentList) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

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

	identity := c.ctxBundle.Identity

	if c.component.State == nil {
		c.component.State = make(map[string]interface{})
	}

	c.component.State["sitePreviewVisible"] = false
	c.component.State["siteAddDrawerVisible"] = false
	c.component.State["siteFormModalVisible"] = false
	c.component.State["formClear"] = false
	c.component.State["siteID"] = 0

	if strings.HasPrefix(event.Operation.String(), apistructs.EdgeOperationChangePage) ||
		event.Operation == apistructs.InitializeOperation ||
		event.Operation == apistructs.RenderingOperation {

		err = c.OperateChangePage(orgID, false, identity)
		if err != nil {
			return err
		}

	} else if event.Operation.String() == apistructs.EdgeOperationDelete {
		err = c.OperateDelete(orgID, event.OperationData, identity)
		if err != nil {
			return err
		}
	} else if event.Operation.String() == apistructs.EdgeOperationAdd {
		err = c.OperateReload(event.OperationData, event.Operation.String())
		if err != nil {
			return err
		}
	} else if event.Operation.String() == apistructs.EdgeOperationUpdate {
		err = c.OperateReload(event.OperationData, event.Operation.String())
		if err != nil {
			return err
		}
	}

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
			{Title: "站点名称", DataIndex: "siteName", Width: 150},
			{Title: "节点数量", DataIndex: "nodeNum", Width: 150},
			{Title: "关联集群", DataIndex: "relatedCluster", Width: 150},
			{Title: "操作", DataIndex: "operate", Width: 150},
		},
	}
}

func getSiteItemOperate(info apistructs.EdgeSiteInfo) apistructs.EdgeItemOperations {
	return apistructs.EdgeItemOperations{
		RenderType: "tableOperation",
		Operations: map[string]apistructs.EdgeItemOperation{
			apistructs.EdgeOperationUpdate: {
				ShowIndex: 2,
				Key:       apistructs.EdgeOperationUpdate,
				Text:      "编辑",
				Reload:    true,
				Meta: map[string]interface{}{
					"id": info.ID,
				},
				Command: apistructs.EdgeJumpCommand{
					Key:    "set",
					Target: "siteFormModal",
					State: apistructs.EdgeJumpCommandState{
						Visible: true,
					},
				},
			},
			apistructs.EdgeOperationAdd: {
				ShowIndex: 1,
				Key:       apistructs.EdgeOperationAdd,
				Text:      "添加节点",
				Reload:    true,
				Meta: map[string]interface{}{
					"id": info.ID,
				},
				Command: apistructs.EdgeJumpCommand{
					Key:    "add",
					Target: "siteAddDrawer",
					State: apistructs.EdgeJumpCommandState{
						Visible: true,
					},
				},
			},
			apistructs.EdgeOperationDelete: {
				ShowIndex:   3,
				Key:         apistructs.EdgeOperationDelete,
				Text:        "删除",
				Confirm:     "此操作将会删除该站点下所有配置项，是否确认删除",
				Disabled:    false,
				DisabledTip: "无法删除",
				Reload:      true,
				Meta: map[string]interface{}{
					"id": info.ID,
				},
			},
		},
	}
}

func renderSiteName(clusterName, name string, id int64) apistructs.EdgeItemOperations {
	return apistructs.EdgeItemOperations{
		RenderType: "linkText",
		Value:      name,
		Operations: map[string]apistructs.EdgeItemOperation{
			"click": {
				Key:    "gotoMachine",
				Reload: false,
				Command: apistructs.EdgeJumpCommand{
					Key:     "goto",
					Target:  "edgeSiteMachine",
					JumpOut: false,
					State: apistructs.EdgeJumpCommandState{
						Params: map[string]interface{}{
							"id": id,
						},
						Query: map[string]interface{}{
							"siteName":    name,
							"clusterName": clusterName,
						},
					},
				},
			},
		},
	}
}
