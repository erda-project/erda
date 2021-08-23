// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orgImage

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func RenderCreator() protocol.CompRender {
	return &OrgImage{}
}

type OrgImage struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  props  `json:"props"`
}

type props struct {
	Src     string `json:"src"`
	Visible bool   `json:"visible"`
	Size    string `json:"size"`
}

type StyleNames struct {
	Normal bool `json:"normal"`
}

func (this *OrgImage) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *OrgImage) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	this.Type = "Image"
	this.Props.Size = "normal"
	this.Props.Visible = true
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}

	this.Props.Src = "frontImg_default_org_img"
	if this.ctxBdl.Identity.OrgID != "" {
		orgDTO, err := this.ctxBdl.Bdl.GetOrg(this.ctxBdl.Identity.OrgID)
		if err != nil {
			return err
		}
		if orgDTO == nil {
			return fmt.Errorf("can not get org")
		}
		if orgDTO.Logo != "" {
			this.Props.Src = orgDTO.Logo
		}
	}

	return nil
}
