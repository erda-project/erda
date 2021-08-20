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

package appsitebreadcrumb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-app-site-ip/i18n"
)

type EdgeAppSiteIPInParam struct {
	ID       int64  `json:"id"`
	AppName  string `json:"appName"`
	SiteName string `json:"siteName"`
}

func (c ComponentBreadCrumb) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	if err := c.SetComponent(component); err != nil {
		return err
	}

	inParam := EdgeAppSiteIPInParam{}

	jsonData, err := json.Marshal(c.ctxBundle.InParams)
	if err != nil {
		return fmt.Errorf("marshal id from inparams error: %v", err)
	}

	if err = json.Unmarshal(jsonData, &inParam); err != nil {
		return fmt.Errorf("unmarshal inparam to object error: %v", err)
	}

	c.component.Data = map[string]interface{}{
		"list": []map[string]interface{}{
			{
				"key":  "appName",
				"item": i18nLocale.Get(i18n.I18nKeySiteList),
			},
			{
				"key":  "siteName",
				"item": inParam.SiteName,
			},
		},
	}

	c.component.Operations = getOperations(inParam.SiteName, inParam.AppName)
	return nil
}

func getOperations(siteName, appName string) apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		apistructs.EdgeOperationClick: apistructs.EdgeOperation{
			Key:      "selectItem",
			FillMeta: "key",
			Meta: map[string]interface{}{
				"key": "",
			},
			Reload: false,
			Command: apistructs.EdgeJumpCommand{
				Key:     "goto",
				Target:  "ecpAppSiteManage",
				JumpOut: false,
				State: apistructs.EdgeJumpCommandState{
					Params: map[string]interface{}{
						"siteName": siteName,
					},
					Query: map[string]interface{}{
						"appName": appName,
					},
				},
			},
		},
	}
}
