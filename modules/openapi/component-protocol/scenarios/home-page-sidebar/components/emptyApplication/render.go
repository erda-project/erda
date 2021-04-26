package emptyApplication

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
	"strconv"
)

func RenderCreator() protocol.CompRender {
	return &EmptyApplication{}
}

type EmptyApplication struct {
	ctxBdl protocol.ContextBundle
	Type string `json:"type"`
	Props props `json:"props"`
	State State `json:"state"`
}

type props struct {
	Visible bool `json:"visible"`
	Tip string `json:"tip"`
	Relative bool `json:"relative"`
}

type State struct {
	ProsNum int `json:"prosNum"`
	AppsNum int `json:"appsNum"`
}

func (this *EmptyApplication) GenComponentState(c *apistructs.Component) error {
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

func (this *EmptyApplication) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *EmptyApplication) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.GenComponentState(c); err != nil {
		return err
	}
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	this.Type = "EmptyHolder"
	if this.State.ProsNum > 0 && this.State.AppsNum == 0 {
		this.Props.Visible = true
	}
	this.Props.Tip = "未加入任何应用"
	this.Props.Relative = true
	return nil
}

func (this *EmptyApplication) getProjectsNum(orgID string) (int, error) {
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
