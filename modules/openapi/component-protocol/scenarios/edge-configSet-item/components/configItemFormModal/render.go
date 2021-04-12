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

package configitemformmodal

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const (
	CfgItemKeyMatchPattern = "^[a-zA-Z-._][a-zA-Z0-9-._]*$"
	CfgItemKeyRegexpError  = "可输入英文字母、数字、中划线、下划线或点, 不能以数字开头"
)

var (
	CfgItemKeyRegexp = fmt.Sprintf("/%v/", CfgItemKeyMatchPattern)
)

func (c ComponentFormModal) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var (
		inParam = apistructs.EdgeRenderingID{}
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

	identity := c.ctxBundle.Identity

	jsonData, err := json.Marshal(c.ctxBundle.InParams)
	if err != nil {
		return fmt.Errorf("marshal id from inparams error: %v", err)
	}

	if err = json.Unmarshal(jsonData, &inParam); err != nil {
		return fmt.Errorf("unmarshal inparam to object error: %v", err)
	}

	if event.Operation.String() == apistructs.EdgeOperationSubmit {
		err = c.OperateSubmit(inParam.ID, identity)
		if err != nil {
			return err
		}
	} else if event.Operation == apistructs.RenderingOperation {
		err = c.OperateRendering(orgID, inParam.ID, identity)
		if err != nil {
			return fmt.Errorf("rendering operation error: %v", err)
		}
		c.component.Operations = getOperations()
		return nil
	}

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

func getProps(sites []map[string]interface{}, modeEdit bool) apistructs.EdgeFormModalProps {
	edgeFormModal := apistructs.EdgeFormModalProps{
		Title: "新建配置项",
		Fields: []apistructs.EdgeFormModalField{
			{
				Key:       "key",
				Label:     "键",
				Component: "input",
				Disabled:  modeEdit,
				Required:  true,
				Rules: []apistructs.EdgeFormModalFieldRule{
					{
						Pattern: CfgItemKeyRegexp,
						Message: CfgItemKeyRegexpError,
					},
				},
				ComponentProps: map[string]interface{}{
					"maxLength": apistructs.EdgeDefaultNameMaxLength,
				},
			},
			{
				Key:       "value",
				Label:     "值",
				Component: "input",
				Required:  true,
				ComponentProps: map[string]interface{}{
					"maxLength": apistructs.EdgeDefaultLagerLength,
				},
			},
			{
				Key:          "scope",
				Label:        "范围",
				Component:    "select",
				Disabled:     modeEdit,
				Required:     true,
				InitialValue: ScopeCommon,
				ComponentProps: map[string]interface{}{
					"placeholder": "请选择范围",
					"options": []map[string]interface{}{
						{"name": "通用", "value": ScopeCommon},
						{"name": "站点", "value": ScopeSite},
					},
				},
			},
			{
				Key:       "sites",
				Label:     "站点",
				Component: "select",
				Disabled:  modeEdit,
				Required:  true,
				ComponentProps: map[string]interface{}{
					"mode":        "multiple",
					"placeholder": "请选择范围",
					"options":     sites,
					"selectAll":   true,
				},
				RemoveWhen: [][]map[string]interface{}{
					{
						{"field": "scope", "operator": "=", "value": ScopeCommon},
					},
				},
			},
		},
	}

	if modeEdit {
		edgeFormModal.Title = "编辑配置项"
	}

	return edgeFormModal
}
