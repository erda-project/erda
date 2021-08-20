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

package cpuChart

import (
	"context"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common/filter"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common/chart"
)

var (
	defaultDuration = 24 * time.Hour
	sqlStatement    = `SELECT cpu_usage_active, timestamp FROM status_page 
	WHERE cluster_name::tag=$cluster_name && hostname::tag=$hostname 
	ORDER BY TIMESTAMP DESC`
)

func (cht *CpuChart) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	cht.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	var (
		resp      *pb.QueryWithInfluxFormatResponse
		nodesResp *apistructs.SteveCollection
		nodesReq  *apistructs.SteveRequest
		nodes     []apistructs.SteveResource
		names     []string
		err       error
	)
	err = common.Transfer(c.State, &cht.State)
	if err != nil {
		return err
	}
	orgId := cht.CtxBdl.Identity.OrgID
	userId := cht.CtxBdl.Identity.UserID
	clusterName := cht.CtxBdl.InParams["clusterName"].(string)
	nodesReq = &apistructs.SteveRequest{
		UserID:      orgId,
		OrgID:       userId,
		Type:        apistructs.K8SNode,
		ClusterName: clusterName,
	}
	nodesResp, err = cht.CtxBdl.Bdl.ListSteveResource(nodesReq)
	if err != nil {
		return err
	}
	for _, res := range nodesResp.Data {
		nodes = append(nodes, res)
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
	case apistructs.CMPDashboardFilterOperationKey:
		for i := len(nodes) - 1; i >= 0; i-- {
			node := nodes[i]
		Next:
			for k, _ := range node.Metadata.Labels {
				for _, f := range filter.PropsInstance.Fields {
					for _, opt := range f.Options {
						if strings.Contains(k, opt.Value) {
							names = append(names, node.ID)
							break Next
						}
					}
				}
			}
		}
	}

	req := apistructs.MetricsRequest{
		ClusterName:  clusterName,
		HostName:     names,
		ResourceType: v1.ResourceCPU,
		OrgID:        orgId,
		UserID:       userId,
	}
	if resp, err = cht.CtxBdl.Bdl.GetMetrics(req); err != nil {
		return err
	}
	items := getDataItem(resp)
	cht.Data = chart.ChartData{Results: chart.Result{Data: items}}
	return nil
}

func getDataItem(response *pb.QueryWithInfluxFormatResponse) []chart.ChartDataItem {
	v := response.Results[0].Series[0].Rows[0].Values
	mem_used := v[0].GetNumberValue()
	mem_available := v[1].GetNumberValue()
	mem_free := v[2].GetNumberValue()
	mem_total := v[3].GetNumberValue()
	return []chart.ChartDataItem{{
		Value: mem_used + mem_available,
		Name:  chart.Distributed_Desc,
	}, {
		Value: mem_free,
		Name:  chart.Free_Desc,
	}, {
		Value: mem_total - mem_free - mem_available - mem_available,
		Name:  chart.Locked_Desc,
	}}
}
func getProps() chart.Props {
	return chart.Props{
		ChartType:  "pie",
		LegendData: []string{"剩余分配", "已分配", "不可分配"},
	}
}
func RenderCreator() protocol.CompRender {
	cc := &CpuChart{}
	cc.Type = "Chart"
	cc.Props = getProps()
	return cc
}
