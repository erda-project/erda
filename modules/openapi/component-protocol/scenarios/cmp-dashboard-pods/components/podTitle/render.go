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

package podTitle

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/sirupsen/logrus"
)

func (pt *PodTitle) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	pt.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	switch event.Operation {
	case apistructs.InitializeOperation:
		err := pt.RenderTitle()
		if err != nil {
			return err
		}
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	if err := pt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

// RenderTitle set data
func (pt *PodTitle) RenderTitle() error {
	var (
		collects *apistructs.SteveCollection
		pods     *apistructs.SteveCollection
		err      error
		orgID    uint64
		clusters []apistructs.ClusterInfo
		cnt      int
	)
	orgID, err = strconv.ParseUint(pt.CtxBdl.Identity.OrgID, 10, 64)
	clusters, err = pt.CtxBdl.Bdl.ListClusters("", orgID)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		req := &apistructs.SteveRequest{
			ClusterName: cluster.Name,
			OrgID:       pt.CtxBdl.Identity.OrgID,
			Type:        apistructs.K8SNode,
		}
		if collects, err = pt.CtxBdl.Bdl.ListSteveResource(req); err != nil {
			return err
		}
		for _, node := range collects.Data {
			nodeReq := &apistructs.SteveRequest{
				ClusterName: cluster.Name,
				OrgID:       pt.CtxBdl.Identity.OrgID,
				Type:        apistructs.K8SPod,
				Name:        node.ID,
			}
			if pods, err = pt.CtxBdl.Bdl.ListSteveResource(nodeReq); err != nil {
				return err
			}
			cnt += len(pods.Data)
		}
	}
	pt.setData(cnt)
	return nil
}

// SetComponentValue transfer CpuInfoTable struct to Component
func (pt *PodTitle) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(pt); err != nil {
		return err
	}
	c.State = state
	return nil
}

// setData assemble rowItem of table
func (pt *PodTitle) setData(cnt int) {
	pt.Props.Size = "small"
	pt.Props.Title = fmt.Sprintf("Pod 总数 %d", cnt)
	pt.Type = "Title"
}
func RenderCreator() protocol.CompRender {
	return &PodTitle{
		Type:  "title",
		Props: Props{},
	}
}
