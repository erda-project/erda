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

package nodeDetail

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
)

// GenComponentState 获取state
func (nd *NodeDetail) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state common.State
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
	nd.State = state
	return nil
}
func (nd *NodeDetail) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	nd.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err := nd.GenComponentState(c);err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.CMPDashboardNodeDetail:
		if err := nd.RenderDetail(); err != nil {
			return err
		}
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	if err := nd.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}
// RenderDetail set data
func (nd *NodeDetail) RenderDetail() error {
	var (
		resp *apistructs.SteveResource
		node *v1.Node
		err error
	)
	req := apistructs.SteveRequest{
		Type:         apistructs.K8SNode,
		ClusterName:   nd.State.ClusterName,
		Name:          nd.State.Name,
	}
	if resp, err = nd.CtxBdl.Bdl.GetSteveResource(&req);err != nil {
		return err
	}
	if err = common.Transfer(resp, node);err != nil{
		return err
	}
	return nd.setData(node)
}

// SetComponentValue transfer CpuInfoTable struct to Component
func (nd *NodeDetail) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(nd.State); err != nil {
		return err
	}
	c.State = state
	return nil
}

// setData assemble rowItem of table
func (nd *NodeDetail) setData(node *v1.Node) error {
	nd.NodeInfo = node.Status.NodeInfo
	return nil
}


