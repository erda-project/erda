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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// GenComponentState 获取state
func (i *ComponentFilter) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
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
	i.State.Conditions = []StateCondition{
		{
			Key:         "title",
			Label:       "标题",
			EmptyText:   "全部",
			Fixed:       true,
			ShowIndex:   2,
			Placeholder: "搜索",
			Type:        "input",
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
	i.ctxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = i.GenComponentState(c); err != nil {
		return
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
		err := i.getFilterOptions()
		if err != nil {
			return err
		}
	case apistructs.CMPDashboardFilterOperationKey:
		i.State.IsFirstFilter = true
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}

	i.SetComponentValue()

	return i.RenderProtocol(c, gs)
}
func (i *ComponentFilter) getFilterOptions() error {
	//var (
	//	clusterList []apistructs.ClusterInfo
	//	ops         []Options
	//	err         error
	//)
	//
	//if clusterList, err = i.ctxBdl.Bdl.ListClusters(""); err != nil{
	//	return nil
	//}
	//for _, cluster := range clusterList {

	//req := apistructs.K8SResourceRequest{
	//	ClusterName:   cluster.Name,
	//	Namespace:     "",
	//	LabelSelector: nil,
	//}
	//if nodes, err := i.ctxBdl.Bdl.ListNodes(&req); err != nil {
	//	return nil
	//} else {
	//for _, node := range nodes {
	//ops = append(ops, node.)
	//}
	//}

	//}
	return nil
}
func getFilterState() State {
	sc := StateCondition{
		Key:         "q",
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
