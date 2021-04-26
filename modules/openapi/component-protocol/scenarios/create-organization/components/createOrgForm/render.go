package createOrgForm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
)

type CreateOrgForm struct {
	ctxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	State      State                  `json:"state"`
	Operations map[string]interface{} `json:"operations"`
}

type SubmitOperation struct {
	Key string `json:"key"`
	Reload bool `json:"reload"`
}

type Props struct {
	Visible bool `json:"visible"`
	Fields []interface{} `json:"fields"`
	ReadOnly bool `json:"readOnly"`
}

type State struct {
	FormData FormData `json:"formData"`
}

type FormData struct {
	Logo        string `json:"logo"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Desc        string `json:"desc"`
	Locale      string `json:"locale"`
	// 创建组织时作为admin的用户id列表
	Admins []string `json:"admins"`

	// 发布商名称
	PublisherName string `json:"publisherName"`
}

type CancelOperation struct {
	Reload bool `json:"reload"`
	Key string `json:"key"`
	Command struct{
		Key string `json:"key"`
		Target string `json:"target"`
		State struct{
			Visible bool `json:"visible"`
		} `json:"state"`
	} `json:"command"`
}

func (o *CreateOrgForm) setProps() {
	o.Props.Visible = true
	o.Props.ReadOnly = false
	fields := make([]interface{}, 0)
	fields = append(fields, map[string]interface{}{
		"label": "组织名称",
		"component": "input",
		"required": true,
		"key": "组织名称",
	})
	fields = append(fields, map[string]interface{}{
		"label": "组织域名",
		"component": "input",
		"required": true,
		"key": "组织域名",
		"componentProps": map[string]interface{}{
			"addonBefore": "erda://",
		},
	})
	fields = append(fields, map[string]interface{}{
		"label": "备注",
		"component": "textarea",
		"key": "备注",
		"componentProps": map[string]interface{}{
			"autoSize": map[string]interface{}{
				"minRows": 4,
				"maxRows": 8,
			},
		},
	})
	fields = append(fields, map[string]interface{}{
		"label": "谁可以看到该组织",
		"component": "radio",
		"key": "谁可以看到该组织",
		"required": true,
		"componentProps": map[string]interface{}{
			"radioType": "radio",
			"displayDesc": true,
		},
		"dataSource": map[string]interface{}{
			"static": []interface{}{
				map[string]interface{}{
					"name": "私人的",
					"desc": "小组及项目只能由成员查看",
					"value": "private",
				},
				map[string]interface{}{
					"name": "公开的",
					"desc": "无需任何身份验证即可查看该组织和任何公开项目",
					"value": "public",
				},
			},
		},
	})
	fields = append(fields, map[string]interface{}{
		"label": "组织图标",
		"component": "upload",
		"key": "组织图标",
		"componentProps": map[string]interface{}{
			"uploadText": "上传图片",
			"sizeLimit": 2,
			"supportFormat": []string{"png", "jpg", "jpeg", "gif", "bmp"},
		},
	})
	o.Props.Fields = fields
}

// GenComponentState 获取state
func (this *CreateOrgForm) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	this.State = state
	return nil
}

func (this *CreateOrgForm) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

//func (o *CreateOrgForm) createOrg() error {
//	req := apistructs.OrgCreateRequest{
//		Logo: o.State.FormData.Logo,
//		Name: o.State.FormData.Name,
//		DisplayName: o.State.FormData.DisplayName,
//		Desc: o.State.FormData.Desc,
//		Locale: o.State.FormData.Locale,
//		Admins: o.State.FormData.Admins,
//		PublisherName: o.State.FormData.PublisherName,
//	}
//	if err := o.ctxBdl.Bdl.org
//}

func (o *CreateOrgForm) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := o.GenComponentState(c); err != nil {
		return err
	}
	if err := o.SetCtxBundle(ctx); err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
		o.setProps()
	case apistructs.OnSubmit:
		o.setProps()
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &CreateOrgForm{}
}