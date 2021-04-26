package createProjectLink

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
	"strconv"
)

type CreateProjectLink struct {
	ctxBdl protocol.ContextBundle
	Type string `json:"type"`
	Props Props `json:"props"`
	Operations map[string]Operation `json:"operations"`
	State State `json:"state"`
}

type Props struct {
	Visible bool `json:"visible"`
	//RenderType string `json:"renderType"`
	//Value map[string]interface{} `json:"value"`
	Text string `json:"text"`
	Disabled bool `json:"disabled"`
	DisabledTip string `json:"disabledTip"`
	Type string `json:"type"`
}

type State struct {
	ProsNum int `json:"prosNum"`
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

func (this *CreateProjectLink) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *CreateProjectLink) GenComponentState(c *apistructs.Component) error {
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

func (this *CreateProjectLink) getProjectsNum(orgID string) (int, error) {
	orgIntId, err := strconv.Atoi(orgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ProjectListRequest{
		OrgID: uint64(orgIntId),
		PageNo: 1,
		PageSize: 1,
	}

	projectDTO, err := this.ctxBdl.Bdl.ListMyProject(this.ctxBdl.Identity.UserID, req)
	if err != nil {
		return 0, err
	}
	if projectDTO == nil {
		return 0, nil
	}
	return projectDTO.Total, nil
}

func (p *CreateProjectLink) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := p.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := p.GenComponentState(c); err != nil {
		return err
	}
	if p.ctxBdl.Identity.OrgID == "" {
		p.Props.Visible = false
		return nil
	}
	orgIntId, err := strconv.Atoi(p.ctxBdl.Identity.OrgID)
	if err != nil {
		return err
	}
	req := &apistructs.PermissionCheckRequest{
		UserID:   p.ctxBdl.Identity.UserID,
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgIntId),
		Resource: apistructs.ProjectResource,
		Action:   apistructs.CreateAction,
	}
	permissionRes, err := p.ctxBdl.Bdl.CheckPermission(req)
	if err != nil {
		return err
	}
	if permissionRes == nil {
		return fmt.Errorf("can not check permission for create project")
	}

	p.Type = "Button"
	var visible bool
	if permissionRes.Access {
		visible = true
	}

	p.Props.Visible = visible
	p.Props.Disabled = false
	p.Props.DisabledTip = "暂无创建项目权限"
	p.Props.Type = "link"
	p.Props.Text = "创建"

	p.Operations = map[string]Operation{
		"click": {
			Command: Command{
				Key: "goto",
				Target: "createProject",
				JumpOut: false,
				Visible: true,
			},
			Key: "click",
			Reload: false,
			Show: false,
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &CreateProjectLink{}
}