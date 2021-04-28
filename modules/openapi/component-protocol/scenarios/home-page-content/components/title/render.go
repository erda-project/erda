// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package title

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

//const (
//	ExpireTypeExpired        ExpireType = "Expired"
//	ExpireTypeExpireIn1Day   ExpireType = "ExpireIn1Day"
//	ExpireTypeExpireIn7Days  ExpireType = "ExpireIn7Days"
//	ExpireTypeExpireIn30Days ExpireType = "ExpireIn30Days"
//)

type Title struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
	State  State  `json:"state"`
}

type Props struct {
	Visible bool   `json:"visible"`
	Title   string `json:"title"`
	Level   int    `json:"level"`
	//TitleStyles TitleStyles `json:"titleStyles"`
	Subtitle       string `json:"subtitle"`
	NoMarginBottom bool   `json:"noMarginBottom"`
	Size           string `json:"size"`
}

type TitleStyles struct {
	FontSize string `json:"fontSize"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (this *Title) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

// GenComponentState 获取state
func (this *Title) GenComponentState(c *apistructs.Component) error {
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

func (t *Title) getUndoneIssueNum() (int, error) {
	orgIDInt, err := strconv.ParseUint(t.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return 0, err
	}

	req := apistructs.IssuePagingRequest{
		PageNo:   1,
		PageSize: 1,
		OrgID:    int64(orgIDInt),
		IssueListRequest: apistructs.IssueListRequest{
			Creators: []string{t.ctxBdl.Identity.UserID},
			StateBelongs: []apistructs.IssueStateBelong{
				apistructs.IssueStateBelongOpen,
				apistructs.IssueStateBelongWorking,
				apistructs.IssueStateBelongReopen,
				apistructs.IssueStateBelongResloved,
				apistructs.IssueStateBelongWontfix,
			},
			External: true,
		},
	}
	req.UserID = t.ctxBdl.Identity.UserID
	issusDTO, err := t.ctxBdl.Bdl.GetIssuesForWorkbench(req)
	if err != nil {
		return 0, err
	}
	if issusDTO == nil {
		return 0, fmt.Errorf("can not get issue response")
	}
	return int(issusDTO.Data.Total), nil
}

func (t *Title) setProps(unDoneIssueNum int) {
	t.Props.Level = 1
	t.Props.Title = "事件"
	t.Props.NoMarginBottom = true
	//t.Props.Subtitle = fmt.Sprintf("你未完成的事项 %d 条", unDoneIssueNum)
	//t.Props.TitleStyles.FontSize = "24px"
	t.Props.Size = "big"
}

func (t *Title) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := t.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := t.GenComponentState(c); err != nil {
		return err
	}
	t.Props.Visible = true
	if t.State.ProsNum == 0 {
		t.setProps(0)
	} else {
		//unDoneIssueNum, err := t.getUndoneIssueNum()
		//if err != nil {
		//	return err
		//}
		//t.setProps(unDoneIssueNum)
		t.setProps(0)
	}
	if t.ctxBdl.Identity.OrgID == "" {
		t.Props.Subtitle = ""
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &Title{}
}
