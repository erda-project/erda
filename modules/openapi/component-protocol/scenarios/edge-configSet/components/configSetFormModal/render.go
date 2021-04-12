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

package configsetformmodal

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func (c ComponentFormModal) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
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

	if event.Operation.String() == apistructs.EdgeOperationSubmit {
		err = c.OperateSubmit(orgID, identity)
		if err != nil {
			return fmt.Errorf("operation submit error: %v", err)
		}
	}

	edgeClusters, err := c.ctxBundle.Bdl.ListEdgeCluster(uint64(orgID), apistructs.EdgeListValueTypeID, identity)
	if err != nil {
		return fmt.Errorf("get avaliable edge clusters error: %v", err)
	}

	c.component.Props = getProps(edgeClusters)
	c.component.Operations = getOperations()
	c.component.State = map[string]interface{}{
		"visible": false,
	}

	return nil
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"submit": apistructs.EdgeOperation{
			Key:    "submit",
			Reload: true,
		},
	}
}

func getProps(cluster []map[string]interface{}) apistructs.EdgeFormModalProps {
	return apistructs.EdgeFormModalProps{
		Title: "新建配置集",
		Fields: []apistructs.EdgeFormModalField{
			{
				Key:       "name",
				Label:     "名称",
				Component: "input",
				Required:  true,
				Rules: []apistructs.EdgeFormModalFieldRule{
					{
						Pattern: apistructs.EdgeDefaultRegexp,
						Message: apistructs.EdgeDefaultRegexpError,
					},
				},
				ComponentProps: map[string]interface{}{
					"maxLength": apistructs.EdgeDefaultNameMaxLength,
				},
			},
			{
				Key:       "cluster",
				Label:     "集群",
				Component: "select",
				Required:  true,
				ComponentProps: map[string]interface{}{
					"placeholder": "请选择关联集群",
					"options":     cluster,
				},
			},
		},
	}
}
