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

	this.Props.Src = "/images/resources/org.png"
	if this.ctxBdl.Identity.OrgID != "" {
		orgDTO, err := this.ctxBdl.Bdl.GetOrg(this.ctxBdl.Identity.OrgID)
		if err != nil {
			return err
		}
		if orgDTO == nil {
			return fmt.Errorf("can not get org")
		}
		if orgDTO.Logo != "" {
			this.Props.Src = fmt.Sprintf("https:%s", orgDTO.Logo)
		}
	}

	return nil
}
