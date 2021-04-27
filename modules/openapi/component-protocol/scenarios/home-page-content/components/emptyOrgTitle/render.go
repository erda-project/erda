package emptyOrgTitle

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EmptyOrgTitle struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}

type Props struct {
	Visible bool   `json:"visible"`
	Title   string `json:"title"`
	Level   int    `json:"level"`
}

func (this *EmptyOrgTitle) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (e *EmptyOrgTitle) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}
	e.Type = "Title"
	if e.ctxBdl.Identity.OrgID == "" {
		e.Props.Visible = true
	}
	e.Props.Title = "你已经是 Erda Cloud 组织的成员"
	e.Props.Level = 2
	return nil
}

func RenderCreator() protocol.CompRender {
	return &EmptyOrgTitle{}
}
