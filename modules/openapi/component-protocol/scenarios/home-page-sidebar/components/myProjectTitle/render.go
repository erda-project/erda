package myProjectTitle

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type MyProjectTitle struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
	State  State  `json:"state"`
}

type Props struct {
	Visible        bool   `json:"visible"`
	Title          string `json:"title"`
	Level          int    `json:"level"`
	NoMarginBottom bool   `json:"noMarginBottom"`
	ShowDivider    bool   `json:"showDivider"`
	Size           string `json:"size"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (this *MyProjectTitle) GenComponentState(c *apistructs.Component) error {
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

func (this *MyProjectTitle) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *MyProjectTitle) getProjectsNum(orgID string) (int, error) {
	orgIntId, err := strconv.Atoi(orgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgIntId),
		PageNo:   1,
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

func (t *MyProjectTitle) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := t.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := t.GenComponentState(c); err != nil {
		return err
	}
	t.Type = "Title"
	t.Props.Title = "项目"
	t.Props.NoMarginBottom = true
	t.Props.Level = 1
	t.Props.Visible = true
	t.Props.ShowDivider = true
	t.Props.Size = "normal"
	return nil
}

func RenderCreator() protocol.CompRender {
	return &MyProjectTitle{}
}
