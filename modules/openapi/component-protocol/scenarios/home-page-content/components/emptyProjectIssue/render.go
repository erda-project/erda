package emptyProjectIssue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
	"strconv"
)

type EmptyProjectIssue struct {
	ctxBdl protocol.ContextBundle
	Type string `json:"type"`
	Props Props `json:"props"`
	State State `json:"state"`
}

type Props struct {
	Tip string `json:"tip"`
	Visible bool `json:"visible"`
	Relative bool `json:"relative"`
	WhiteBg bool `json:"whiteBg"`
	PaddingY bool `json:"paddingY"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (this *EmptyProjectIssue) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *EmptyProjectIssue) getProjectsNum(orgID string) (int, error) {
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

func (this *EmptyProjectIssue) GenComponentState(c *apistructs.Component) error {
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

func (e *EmptyProjectIssue) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := e.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := e.GenComponentState(c); err != nil {
		return err
	}

	e.Type = "EmptyHolder"
	e.Props.Tip = "已加入的项目中，无待完成事项"
	e.Props.WhiteBg = true
	e.Props.Relative = true
	e.Props.PaddingY = true
	if e.ctxBdl.Identity.OrgID != "" && e.State.ProsNum == 0 {
		prosNum, err := e.getProjectsNum(e.ctxBdl.Identity.OrgID)
		if err != nil {
			return err
		}
		if prosNum != 0 {
			e.Props.Visible = true
		}
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &EmptyProjectIssue{}
}