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

package configsetlist

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EdgeConfigsetItem struct {
	ID             int64                         `json:"id"`
	ConfigsetName  string                        `json:"configSetName"`
	RelatedCluster string                        `json:"relatedCluster"`
	Operate        apistructs.EdgeItemOperations `json:"operate"`
}

func (c *ComponentConfigsetList) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

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

	if strings.HasPrefix(event.Operation.String(), apistructs.EdgeOperationChangePage) ||
		event.Operation == apistructs.InitializeOperation ||
		event.Operation == apistructs.RenderingOperation {
		err = c.OperateChangePage(orgID, false, identity)
		if err != nil {
			return fmt.Errorf("operation change page error: %v", err)
		}
	} else if event.Operation.String() == apistructs.EdgeOperationDelete {
		err = c.OperateDelete(orgID, event.OperationData, identity)
		if err != nil {
			return fmt.Errorf("operation delete error: %v", err)
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
			{Title: "配置集名称", DataIndex: "configSetName", Width: 150},
			{Title: "关联集群", DataIndex: "relatedCluster", Width: 150},
			{Title: "操作", DataIndex: "operate", Width: 150},
		},
	}
}

func getConfigsetItem(id int64, configSetName string) apistructs.EdgeItemOperations {
	return apistructs.EdgeItemOperations{
		RenderType: "tableOperation",
		Operations: map[string]apistructs.EdgeItemOperation{
			"viewDetail": {
				ShowIndex: 1,
				Key:       "viewDetail",
				Text:      "详情",
				Reload:    false,
				Meta: map[string]interface{}{
					"id":            id,
					"configSetName": configSetName,
				},
				Command: apistructs.EdgeJumpCommand{
					JumpOut: false,
					Key:     "goto",
					State: apistructs.EdgeJumpCommandState{
						Params: map[string]interface{}{
							"id": id,
						},
						Query: map[string]interface{}{
							"configSetName": configSetName,
						},
					},
					Target: "edgeSettingDetail",
				},
			},
			apistructs.EdgeOperationDelete: {
				ShowIndex:   2,
				Key:         apistructs.EdgeOperationDelete,
				Text:        "删除",
				Confirm:     "此操作会导致配置集下的所有配置删除，是否确认删除",
				Disabled:    false,
				DisabledTip: "无法删除",
				Reload:      true,
				Meta:        map[string]interface{}{"id": id},
			},
		},
	}
}
