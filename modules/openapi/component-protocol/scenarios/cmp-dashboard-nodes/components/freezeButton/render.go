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

package freezeButton

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
)

var buttonOps = map[string]Operation{
	"click": {
		Key:    "freezy",
		Reload: true,
	},
}

var props = Props{
	Text:    "冻结",
	Type:    "primary",
	Tooltip: "只有worker节点可以冻结",
}

func (fb *FreezeButton) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var (
		err  error
		meta Meta
		resp *apistructs.SteveResource
		node = v1.Node{}
	)
	fb.ctxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = common.Transfer(event.OperationData, &meta); err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.CMPDashboardFreezeNode:
		req := apistructs.SteveRequest{
			Type:        apistructs.K8SNode,
			ClusterName: meta.ClusterName,
			Name:        meta.NodeName,
		}
		if resp, err = fb.ctxBdl.Bdl.GetSteveResource(&req); err != nil {
			return err
		}
		err = common.Transfer(resp, &node)
		if err != nil {
			return err
		}
		if v := node.GetLabels()["node-role.kubernetes.io/master"]; v == "true" {
			return common.NodeRoleInvalidErr
		}
		if node.Spec.Unschedulable {
			logrus.Infof("node has already cordoned")
			return nil
		}
		node.Spec.Unschedulable = true
		req.Obj = &node
		if _, err = fb.ctxBdl.Bdl.UpdateSteveResource(&req); err != nil {
			logrus.Errorf("Uncordon node %s/%s error :%v", meta.ClusterName, meta.NodeName, err)
			return err
		}
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	return nil
}
func GetOpsInfo(opsData interface{}) (*Meta, error) {
	if opsData == nil {
		err := fmt.Errorf("empty operation data")
		return nil, err
	}
	var meta Meta
	cont, err := json.Marshal(opsData)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", opsData, err)
		return nil, err
	}
	err = json.Unmarshal(cont, &meta)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	return &meta, nil
}
func getOperation() map[string]Operation {
	return buttonOps
}

func getProps() Props {
	return props
}
func RenderCreator() protocol.CompRender {
	return &FreezeButton{
		Type:       "Button",
		Props:      getProps(),
		Operations: getOperation(),
	}
}
