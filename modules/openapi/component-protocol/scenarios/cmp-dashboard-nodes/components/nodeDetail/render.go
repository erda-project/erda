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

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common/table"
)

func (nd *NodeDetail) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var state table.State
	nd.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err := common.Transfer(c.State, &state); err != nil {
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
		node *apistructs.SteveResource
		err  error
	)
	req := apistructs.SteveRequest{
		Type:        apistructs.K8SNode,
		ClusterName: nd.State.ClusterName,
		Name:        "GET",
	}
	if node, err = nd.CtxBdl.Bdl.GetSteveResource(&req); err != nil {
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
func (nd *NodeDetail) setData(resource *apistructs.SteveResource) error {
	var node v1.Node
	err := common.Transfer(resource, node)
	if err != nil {
		return err
	}
	nd.NodeInfo = node.Status.NodeInfo
	return nil
}
func RenderCreator() protocol.CompRender {
	return &NodeDetail{}
}
