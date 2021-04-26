package erdaLogo

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
)

type ErdaLogo struct {
	ctxBdl protocol.ContextBundle
	Type string `json:"type"`
	Props Props `json:"props"`
}

type Props struct {
	Visible bool `json:"visible"`
	Src string `json:"src"`
	IsCircle bool `json:"isCircle"`
	Size string `json:"size"`
	Type string `json:"type"`
}

type StyleNames struct {
	Small bool `json:"small"`
	Mt8 bool `json:"mt8"`
	Circle bool `json:"circle"`
}

func (this *ErdaLogo) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (e *ErdaLogo) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}
	e.Type = "Image"
	e.Props.Type = "erda"
	if e.ctxBdl.Identity.OrgID == "" {
		e.Props.Visible = true
	}
	//e.Props.Src = "https://ss1.bdstatic.com/70cFuXSh_Q1YnxGkpoWK1HF6hhy/it/u=3355464299,584008140&fm=26&gp=0.jpg"
	e.Props.IsCircle = true
	e.Props.Size = "small"
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ErdaLogo{}
}
