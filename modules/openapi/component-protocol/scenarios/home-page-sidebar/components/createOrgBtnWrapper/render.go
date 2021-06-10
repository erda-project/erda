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

package createOrgBtnWrapper

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/conf"
)

type CreateOrgBtnWrapper struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}

type Props struct {
	ContentSetting string `json:"contentSetting"`
	Visible        bool   `json:"visible"`
}

func (c *CreateOrgBtnWrapper) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	c.ctxBdl = bdl
	return nil
}

func (p *CreateOrgBtnWrapper) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := p.SetCtxBundle(ctx); err != nil {
		return err
	}

	p.Type = "RowContainer"
	p.Props.ContentSetting = "center"
	if p.ctxBdl.Identity.OrgID == "" && conf.CreateOrgEnabled() {
		p.Props.Visible = true
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &CreateOrgBtnWrapper{}
}
