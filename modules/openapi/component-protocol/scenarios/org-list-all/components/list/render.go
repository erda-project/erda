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

package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const defaultOrgImage = "/images/default-org-icon.svg"

func (i *ComponentList) unmarshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state State
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}
	i.State = state
	return nil
}

func (i *ComponentList) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(i.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	c.Data = map[string]interface{}{
		"list": i.Data,
	}
	c.State = state
	c.Props = i.Props
	c.Operations = i.Operations
	return nil
}

func (i *ComponentList) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.CtxBdl = b
	return nil
}

func (i *ComponentList) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if event.Operation != apistructs.InitializeOperation {
		err = i.unmarshal(c)
		if err != nil {
			return err
		}
	}

	defer func() {
		fail := i.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err := i.SetCtxBundle(bdl); err != nil {
		return err
	}

	i.initProperty()
	switch event.Operation {
	case apistructs.InitializeOperation, apistructs.RenderingOperation, apistructs.ChangeOrgsPageNoOperationKey, apistructs.ChangeOrgsPageSizeOperationKey:
		if err := i.RenderPublicOrgs(); err != nil {
			return err
		}
	}
	return nil
}

func (i *ComponentList) RenderPublicOrgs() error {
	if i.State.PageSize == 0 {
		i.State.PageSize = 10
		i.State.PageNo = 1
	}
	if i.State.SearchRefresh {
		i.State.PageNo = 1
	}
	req := apistructs.OrgSearchRequest{
		PageNo:   i.State.PageNo,
		PageSize: i.State.PageSize,
		Q:        i.State.SearchEntry,
	}
	req.UserID = i.CtxBdl.Identity.UserID
	orgs, err := i.CtxBdl.Bdl.ListPublicOrgs(&req)
	if err != nil {
		return err
	}

	req = apistructs.OrgSearchRequest{}
	req.UserID = i.CtxBdl.Identity.UserID
	myOrgs, err := i.CtxBdl.Bdl.ListOrgs(&req)
	if err != nil {
		return err
	}
	orgMap := map[uint64]bool{}
	for _, o := range myOrgs.List {
		orgMap[o.ID] = true
	}

	i.State.Total = orgs.Total
	data := []Org{}
	for _, org := range orgs.List {
		item := Org{
			Id:          org.ID,
			Title:       org.DisplayName,
			Description: org.Desc,
			PrefixImg:   defaultOrgImage,
			ExtraInfos: []ExtraInfo{
				{
					Icon: "earth",
					Text: "公开组织",
				},
			},
		}
		if org.Logo != "" {
			item.PrefixImg = org.Logo
		}

		if _, ok := orgMap[org.ID]; ok {
			item.ExtraInfos = append(item.ExtraInfos, ExtraInfo{
				Icon: "renyuan",
				Text: "已加入",
			})
		}

		item.Operations = map[string]interface{}{
			"click": ClickOperation{
				Key:    "click",
				Show:   false,
				Reload: false,
				Command: Command{
					Key:    "goto",
					Target: "https://" + org.Domain + "/workBench/projects",
				},
			},
		}

		data = append(data, item)
	}
	i.Data = data
	i.State.SearchRefresh = false
	return nil
}

func getOperation(operationData *ExitOperation, event apistructs.ComponentEvent) error {
	if event.OperationData == nil {
		return nil
	}
	b, err := json.Marshal(event.OperationData)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &operationData); err != nil {
		return err
	}
	return nil
}

func (i *ComponentList) initProperty() {
	i.Operations = map[string]interface{}{
		"changePageNo": OperationBase{
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": OperationBase{
			Key:    "changePageSize",
			Reload: true,
		},
	}

	i.Props.PageSizeOptions = []string{"10", "20", "50", "100"}
}

func RenderCreator() protocol.CompRender {
	return &ComponentList{}
}
