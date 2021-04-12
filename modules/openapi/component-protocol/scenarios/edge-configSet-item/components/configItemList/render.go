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

package configitemlist

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EdgeConfigItem struct {
	ID          int64                         `json:"id"`
	ConfigName  string                        `json:"configName"`
	ConfigValue string                        `json:"configValue"`
	SiteName    string                        `json:"siteName"`
	CreateTime  string                        `json:"createTime"`
	UpdateTime  string                        `json:"updateTime"`
	Operate     apistructs.EdgeItemOperations `json:"operate"`
}

func (c *ComponentConfigItemList) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	identity := c.ctxBundle.Identity

	if c.component.State == nil {
		c.component.State = map[string]interface{}{}
	}

	c.component.State["configItemFormModalVisible"] = false
	c.component.State["formClear"] = false
	c.component.State["configSetItemID"] = 0

	if strings.HasPrefix(event.Operation.String(), apistructs.EdgeOperationChangePage) ||
		event.Operation == apistructs.InitializeOperation ||
		event.Operation == apistructs.RenderingOperation {
		err := c.OperateChangePage(false, identity)
		if err != nil {
			return err
		}
	} else if event.Operation.String() == apistructs.EdgeOperationDelete {
		err := c.OperateDelete(event.OperationData, identity)
		if err != nil {
			return err
		}
	} else if event.Operation.String() == apistructs.EdgeOperationUpdate {
		err := c.OperateReload(event.OperationData)
		if err != nil {
			return fmt.Errorf("operation reload error: %v", err)
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
			{Title: "配置项", DataIndex: "configName", Width: 150},
			{Title: "值", DataIndex: "configValue", Width: 150},
			{Title: "站点", DataIndex: "siteName", Width: 150},
			{Title: "创建时间", DataIndex: "createTime", Width: 150},
			{Title: "更新时间", DataIndex: "updateTime", Width: 150},
			{Title: "操作", DataIndex: "operate", Width: 150},
		},
	}
}

func getConfigsetItem(info apistructs.EdgeCfgSetItemInfo) apistructs.EdgeItemOperations {
	formData := map[string]interface{}{
		"id":    info.ID,
		"key":   info.ItemKey,
		"value": info.ItemValue,
	}

	if info.Scope == "public" {
		formData["scope"] = "COMMON"
	} else if info.Scope == "site" {
		formData["scope"] = "SITE"
		formData["sites"] = []string{info.SiteName}
	}

	return apistructs.EdgeItemOperations{
		RenderType: "tableOperation",
		Operations: map[string]apistructs.EdgeItemOperation{
			"viewDetail": {
				ShowIndex: 1,
				Key:       apistructs.EdgeOperationUpdate,
				Text:      "编辑",
				Reload:    true,
				Meta: map[string]interface{}{
					"id": info.ID,
				},
				Command: apistructs.EdgeJumpCommand{
					Key:    "set",
					Target: "configItemFormModal",
					State: apistructs.EdgeJumpCommandState{
						Visible: true,
					},
				},
			},
			apistructs.EdgeOperationDelete: {
				ShowIndex:   2,
				Key:         apistructs.EdgeOperationDelete,
				Text:        "删除",
				Confirm:     "是否确认删除",
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
