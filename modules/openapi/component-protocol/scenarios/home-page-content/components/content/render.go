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

package content

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type Content struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	State  State  `json:"state"`
	Props  Props  `json:"props"`
}

type Props struct {
	Visible bool `json:"visible"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (t *Content) getWorkbenchData() (*apistructs.WorkbenchResponse, error) {
	orgID, err := strconv.ParseUint(t.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return nil, err
	}
	req := apistructs.WorkbenchRequest{
		OrgID:     orgID,
		PageSize:  1,
		PageNo:    1,
		IssueSize: 1,
	}
	res, err := t.ctxBdl.Bdl.GetWorkbenchData(t.ctxBdl.Identity.UserID, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *Content) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *Content) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	this.Type = "Container"
	this.Props.Visible = true
	return nil
}

func RenderCreator() protocol.CompRender {
	return &Content{}
}
