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

package clusterFilter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// SetCtxBundle 设置bundle
func (i *ComponentFilter) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.ctxBdl = b
	return nil
}

// GenComponentState mapping c.State to i.State
func (i *ComponentFilter) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return common.ProtocolComponentEmptyErr
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal components state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal components state failed, content:%v, err:%v", cont, err)
		return err
	}
	i.State = state
	return nil
}

func (i *ComponentFilter) SetComponentValue() {
	i.Props = Props{
		Delay: 1000,
	}
	i.Operations = map[string]interface{}{
		apistructs.CMPDashboardFilterOperationKey.String(): Operation{
			Reload: true,
			Key:    "clusterFilter",
		},
	}
}

// RenderProtocol 渲染组件
func (i *ComponentFilter) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) error {
	stateValue, err := json.Marshal(i.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}
	c.State = state
	c.Data["list"] = i.Options
	c.Props = i.Props
	c.Operations = i.Operations
	return nil
}

func (i *ComponentFilter) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = i.SetCtxBundle(bdl); err != nil {
		return
	}
	if err = i.GenComponentState(c); err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
		ops ,err := i.getFilterOptions()
		if err != nil {
			return err
		}
		i.Options = ops
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}

	i.SetComponentValue()

	return i.RenderProtocol(c, gs)
}
func (i *ComponentFilter) getFilterOptions() ([]Options,error) {
	clusters, err := i.ctxBdl.Bdl.ListClusters("")
	if err != nil {
		return nil,err
	}
	var ops []Options
	for _,cluster := range clusters{
		ops = append(ops, Options{
			Label:    "",
			Value:    cluster.Name,
		})
	}

	return ops,nil
}
func getFilterState() State {
	sc := StateCondition{
		Key:         "filter",
		Label:       "标题",
		Placeholder: "请输入关键字查询",
		Type:        "input",
		Fixed:       true,
	}
	state := State{
		Conditions:    []StateCondition{sc},
		IsFirstFilter: false,
	}
	return state
}
func RenderCreator() protocol.CompRender {
	return &ComponentFilter{
		CommonFilter: CommonFilter{
			Type:       "ContractiveFilter",
			State:      getFilterState(),
			Operations: nil,
		},
	}
}
