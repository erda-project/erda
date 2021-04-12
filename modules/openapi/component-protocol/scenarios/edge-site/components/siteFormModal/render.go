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

package siteformmodal

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const (
	SiteNameMatchPattern = "^[a-z][a-z0-9-]*[a-z0-9]$"
	SitNameRegexpError   = "可输入小写字母、数字、中划线; 必须以小写字母开头, 以小写字母或数字结尾"
)

var (
	SiteNameMatchRegexp = fmt.Sprintf("/%v/", SiteNameMatchPattern)
)

type SiteFormSubmitBase struct {
	ID       int64  `json:"id"`
	SiteName string `json:"siteName"`
	Desc     string `json:"desc"`
}

type SiteFormCreate struct {
	SiteFormSubmitBase
	RelatedCluster int64 `json:"relatedCluster"`
}

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

	if c.component.State == nil {
		c.component.State = map[string]interface{}{}
	}

	identity := c.ctxBundle.Identity

	if event.Operation.String() == apistructs.EdgeOperationSubmit {
		err = c.OperateSubmit(orgID, identity)
		if err != nil {
			return err
		}
	} else if event.Operation == apistructs.RenderingOperation {
		err = c.OperateRendering(orgID, identity)
		if err != nil {
			return err
		}
		c.component.Operations = getOperations()
		return nil
	}

	c.component.State = map[string]interface{}{
		"reload":  true,
		"visible": false,
	}
	c.component.Operations = getOperations()

	return nil
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"submit": map[string]interface{}{
			"key":    "submit",
			"reload": true,
		},
	}
}

func getProps(relatedCluster []map[string]interface{}, modeEdit bool) apistructs.EdgeFormModalProps {

	siteNameField := apistructs.EdgeFormModalField{
		Key:       "siteName",
		Label:     "站点名称",
		Component: "input",
		Required:  true,
		Rules: []apistructs.EdgeFormModalFieldRule{
			{
				Pattern: SiteNameMatchRegexp,
				Message: SitNameRegexpError,
			},
		},
		ComponentProps: map[string]interface{}{
			"maxLength": SiteNameLength,
		},
	}

	clusterField := apistructs.EdgeFormModalField{
		Key:       "relatedCluster",
		Label:     "关联集群",
		Component: "select",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请选择关联集群",
			"options":     relatedCluster,
		},
	}

	descField := apistructs.EdgeFormModalField{
		Key:       "desc",
		Label:     "站点描述",
		Component: "textarea",
		Required:  false,
		ComponentProps: map[string]interface{}{
			"maxLength": apistructs.EdgeDefaultLagerLength / 2,
		},
	}

	siteNameField.Disabled = modeEdit
	clusterField.Disabled = modeEdit

	props := apistructs.EdgeFormModalProps{
		Title: "新建站点",
		Fields: []apistructs.EdgeFormModalField{
			siteNameField,
			clusterField,
			descField,
		},
	}

	if modeEdit {
		props.Title = "编辑站点"
	}

	return props
}
