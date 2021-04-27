package emptyOrganization

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type EmptyOrganization struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}

type Props struct {
	Visible bool `json:"visible"`
}

func (this *EmptyOrganization) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (t *EmptyOrganization) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := t.SetCtxBundle(ctx); err != nil {
		return err
	}
	var visible bool
	if t.ctxBdl.Identity.OrgID == "" {
		visible = true
	}
	t.Type = "Container"
	t.Props.Visible = visible
	return nil
}

func RenderCreator() protocol.CompRender {
	return &EmptyOrganization{}
}
