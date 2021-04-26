package orgImage

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
)

func RenderCreator() protocol.CompRender {
	return &OrgImage{}
}

type OrgImage struct {
	ctxBdl protocol.ContextBundle
	Type string `json:"type"`
	Props props `json:"props"`
	//State State `json:"state"`
}

type props struct {
	Src string `json:"src"`
	Visible bool `json:"visible"`
	Size string `json:"size"`
	Type string `json:"type"`
}

type StyleNames struct {
	Normal bool `json:"normal"`
}

type State struct {
	//OrgID string `json:"orgID"`
	//PrefixImage string `json:"prefixImage"`
}

//func (this *OrgImage) GenComponentState(c *apistructs.Component) error {
//	if c == nil || c.State == nil {
//		return nil
//	}
//	var state State
//	cont, err := json.Marshal(c.State)
//	if err != nil {
//		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
//		return err
//	}
//	err = json.Unmarshal(cont, &state)
//	if err != nil {
//		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
//		return err
//	}
//	this.State = state
//	return nil
//}

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
	this.Props.Type = "organization"
	this.Props.Visible = true
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}

	//this.Props.Src = "https://ss1.bdstatic.com/70cFuXSh_Q1YnxGkpoWK1HF6hhy/it/u=3355464299,584008140&fm=26&gp=0.jpg"
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