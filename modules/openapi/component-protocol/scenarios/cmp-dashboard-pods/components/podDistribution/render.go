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

package podDistribution

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-pods/common"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

func (pd *PodDistribution) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	pd.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err := common.Transfer(c.State, &pd.State)
	if err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		err := pd.RenderPodDistribution()
		if err != nil {
			return err
		}
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}

	if err := pd.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}
func (pd *PodDistribution) RenderPodDistribution() error {
	var (
		podList      []apistructs.SteveResource
		pods         []apistructs.SteveResource
		resp         *apistructs.SteveCollection
		err          error
		filter       string
		orgID        uint64
		clusterNames []apistructs.ClusterInfo
	)
	orgID, err = strconv.ParseUint(pd.CtxBdl.Identity.OrgID, 10, 64)
	if pd.State.ClusterName != "" {
		clusterNames = append([]apistructs.ClusterInfo{}, apistructs.ClusterInfo{Name: pd.State.ClusterName})
	} else {
		clusterNames, err = bdl.Bdl.ListClusters("", orgID)
		if err != nil {
			return err
		}
	}
	// Get all nodes by cluster name
	for _, clusterName := range clusterNames {
		req := &apistructs.SteveRequest{}
		req.Name = clusterName.Name
		req.ClusterName = clusterName.Name
		req.Type = apistructs.K8SPod
		req.OrgID = pd.CtxBdl.Identity.OrgID
		req.UserID = pd.CtxBdl.Identity.UserID
		resp, err = pd.CtxBdl.Bdl.ListSteveResource(req)
		podList = append(podList, resp.Data...)
	}
	if filter == "" {
		pods = podList
	} else {
		// Filter by node name or node uid
		for _, pod := range podList {
			if strings.Contains(pod.Metadata.Name, filter) || strings.Contains(pod.ID, filter) {
				pods = append(pods, pod)
			}
		}
	}
	cnts := make(map[string]int)
	for _, pod := range pods {
		status := v1.PodStatus{}
		err := common.Transfer(pod.Status, &status)
		if err != nil {
			return err
		}
		cnts[string(status.Phase)]++
	}
	pd.SetData(cnts,len(pods))
	return nil
}

func (pd *PodDistribution) SetData(data map[string]int,sum int) {
	i:=0
	d := make([]Data,len(data))
	for k,v:=range data {
		d[i].Label = k
		d[i].Tip = fmt.Sprintf("%d",v)
		d[i].Value= v
		d[i].Color = getColor(v,sum)
	}
	pd.Data["list"] = d
}
func getColor(cnt,sum int)string{
	return ""
}
func (pd *PodDistribution) SetComponentValue(c *apistructs.Component) error {
	c.Data = pd.Data
	return nil
}

func RenderCreator()protocol.CompRender{
	return &PodDistribution{
		Type:   "LinearDistribution",
	}
}
