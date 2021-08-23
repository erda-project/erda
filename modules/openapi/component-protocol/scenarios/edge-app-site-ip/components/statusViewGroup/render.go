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

package statusviewgroup

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-app-site-ip/i18n"
	i18r "github.com/erda-project/erda/pkg/i18n"
)

func (c ComponentViewGroup) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

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

	if event.Operation == apistructs.EdgeOperationChangeRadio || event.Operation == apistructs.InitializeOperation {
		err = c.OperationChangeViewGroup()
		if err != nil {
			return err
		}

		err = c.Operation(orgID, identity)
		if err != nil {
			return err
		}

		c.component.Operations = getOperations()
	}

	return nil
}

func getProps(total, success, error int, lr *i18r.LocaleResource) apistructs.EdgeRadioProps {
	return apistructs.EdgeRadioProps{
		RadioType:   "button",
		ButtonStyle: "outline",
		Size:        "small",
		Options: []apistructs.EdgeButtonOption{
			{Text: fmt.Sprintf("%s(%d)", lr.Get(i18n.I18nKeyAll), total), Key: "total"},
			{Text: fmt.Sprintf("%s(%d)", lr.Get(i18n.I18nKeyRunning), success), Key: "success"},
			{Text: fmt.Sprintf("%s(%d)", lr.Get(i18n.I18nKeyStopped), error), Key: "error"},
		},
	}
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"onChange": apistructs.EdgeOperation{
			Key:    apistructs.EdgeOperationChangeRadio,
			Reload: true,
		},
	}
}
