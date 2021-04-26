package myProject

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
	"strconv"
)

type MyProject struct {
	ctxBdl     protocol.ContextBundle
	Type string `json:"type"`
	Props Props `json:"props"`
	State State `json:"state"`
}

type Props struct {
	Visible bool `json:"visible"`
	SpaceSize string `json:"spaceSize"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (this *MyProject) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *MyProject) getProjectsNum(orgID string) (int, error) {
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

func (t *MyProject) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := t.SetCtxBundle(ctx); err != nil {
		return err
	}
	var err error
	var prosNum int
	if t.ctxBdl.Identity.OrgID == "" {
		prosNum = 0
	} else {
		prosNum, err = t.getProjectsNum(t.ctxBdl.Identity.OrgID)
		if err != nil {
			return err
		}
	}
	t.State.ProsNum = prosNum
	t.Type = "Container"
	t.Props.Visible = true
	t.Props.SpaceSize = "middle"
	return nil
}

func RenderCreator() protocol.CompRender {
	return &MyProject{}
}
