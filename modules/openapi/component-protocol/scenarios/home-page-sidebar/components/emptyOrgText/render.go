package emptyOrgText

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
)

const (
	DefaultFontSize = 16
	DefaultLineHeight = 24
	DefaultType = "TextGroup"
)

func RenderCreator() protocol.CompRender {
	return &EmptyOrgText{}
}

type EmptyOrgText struct {
	ctxBdl protocol.ContextBundle
	Type string `json:"type"`
	Props props `json:"props"`
	Operations map[string]Operation `json:"operations"`
}

type State struct {
	//OrgID string `json:"orgID"`
	//PrefixImage string `json:"prefixImage"`
}

type props struct {
	Visible bool `json:"visible"`
	Align string `json:"align"`
	Value []interface{} `json:"value"`
}

type Command struct {
	Key string `json:"key"`
	Target string `json:"target"`
	JumpOut bool `json:"jumpOut"`
	Visible bool `json:"visible"`
}

type Operation struct {
	Command Command `json:"command"`
	Key string `json:"key"`
	Reload bool `json:"reload"`
	Show bool `json:"show"`
}

//func (this *EmptyOrgText) GenComponentState(c *apistructs.Component) error {
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

func (this *EmptyOrgText) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *EmptyOrgText) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	//if err := this.GenComponentState(c); err != nil {
	//	return err
	//}
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	this.Type = DefaultType
	this.Props.Align = "center"
	var visible bool
	if this.ctxBdl.Identity.OrgID == "" {
		visible = true
	}
	this.Props.Visible = visible

	this.Props.Value = make([]interface{}, 0)
	this.Props.Value = append(this.Props.Value, map[string]interface{}{
		"props": map[string]interface{}{
			"renderType": "text",
			"visible": visible,
			"value": "未加入任何组织",
		},
	})
	this.Props.Value = append(this.Props.Value, map[string]interface{}{
		"props": map[string]interface{}{
			"renderType": "linkText",
			"visible": visible,
			"value": map[string]interface{}{
				"text": []interface{}{map[string]interface{}{
					"text": "了解如何受邀加入到组织",
					"operationKey": "toJoinOrgDoc",
				}},
			},
		},
	})
	this.Props.Value = append(this.Props.Value, map[string]interface{}{
		"props": map[string]interface{}{
			"renderType": "linkText",
			"visible": visible,
			"value": map[string]interface{}{
				"text": []interface{}{map[string]interface{}{
					"text": "浏览公开组织信息",
					"operationKey": "toPublicOrgPage",
				}},
			},
		},
	})
	this.Operations = make(map[string]Operation)
	this.Operations["toJoinOrgDoc"] = Operation{
		Command: Command{
			Key: "goto",
			Target: "https://docs.erda.cloud/",
			JumpOut: true,
			Visible: visible,
		},
		Key: "click",
		Reload: false,
		Show: false,
	}
	this.Operations["toPublicOrgPage"] = Operation{
		Command: Command{
			Key: "goto",
			Target: "orgList",
			JumpOut: true,
			Visible: visible,
		},
	}
	return nil
}


