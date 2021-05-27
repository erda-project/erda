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

package orgLogo

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type OrgLogo struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}

type Props struct {
	Visible  bool   `json:"visible"`
	Src      string `json:"src"`
	IsCircle bool   `json:"isCircle"`
	Size     string `json:"size"`
	//StyleNames StyleNames `json:"styleNames"`
}

type StyleNames struct {
	Small  bool `json:"small"`
	Mt8    bool `json:"mt8"`
	Circle bool `json:"circle"`
}

func (this *OrgLogo) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (e *OrgLogo) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}
	e.Type = "Image"
	e.Props.Src = "/images/resources/org.png"
	e.Props.IsCircle = true
	e.Props.Size = "small"
	if e.ctxBdl.Identity.OrgID != "" {
		orgDTO, err := e.ctxBdl.Bdl.GetOrg(e.ctxBdl.Identity.OrgID)
		if err != nil {
			return err
		}
		if orgDTO == nil {
			return fmt.Errorf("can not get org")
		}
		if orgDTO.Logo != "" {
			e.Props.Src = orgDTO.Logo
		}
		e.Props.Visible = true
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &OrgLogo{}
}
