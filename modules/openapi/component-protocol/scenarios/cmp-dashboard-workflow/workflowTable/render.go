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

package workflowTable

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"google.golang.org/protobuf/types/known/structpb"
	"strconv"
	"strings"
	"time"

	"github.com/cznic/mathutil"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

var tableProperties = map[string]interface{}{
	"rowKey": "id",
	"columns": []Columns{
		{DataIndex: "status", Title: "状态"},
		{DataIndex: "name", Title: "名称"},
		{DataIndex: "namespace", Title: "命名空间"},
		{DataIndex: "image", Title: "镜像"},
		{DataIndex: "type", Title: "类型"},
		{DataIndex: "alive", Title: "存活时间"},
	},
	"bordered":        true,
	"selectable":      true,
}
var metricsServer = servicehub.New().Service("metrics-query").(pb.MetricServiceServer)

func (wt *WorkflowTable) Import(c *apistructs.Component) error {
	var (
		b   []byte
		err error
	)
	if b, err = json.Marshal(c); err != nil {
		return err
	}
	if err = json.Unmarshal(b, wt); err != nil {
		return err
	}
	return nil
}

func (wt *WorkflowTable) Export(c *apistructs.Component, gs *apistructs.GlobalStateData) error {
	// set components data
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, wt); err != nil {
		return err
	}
	return nil
}

// GenComponentState 获取state
func (wt *WorkflowTable) GenComponentState(c *apistructs.Component) error {
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
	wt.State = state
	return nil
}
func (wt *WorkflowTable) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	wt.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	switch event.Operation {
	case apistructs.InitializeOperation:
		if err := wt.listOperationHandler(wt.CtxBdl); err != nil {
			return err
		}
	case apistructs.CMPDashboardChangePageSizeOperationKey:
		if err := wt.RenderChangePageSize(event.OperationData); err != nil {
			return err
		}
	case apistructs.CMPDashboardChangePageNoOperationKey:
		if err := wt.RenderChangePageNo(event.OperationData); err != nil {
			return err
		}
	case apistructs.RenderingOperation:
		// IsFirstFilter delivered from filer component
		if wt.State.IsFirstFilter {
			wt.State.PageNo = 1
			wt.State.IsFirstFilter = false
		}
	case apistructs.ExecuteClickRowNoOperationKey:
		if err := wt.clickRowOperationHandler(wt.CtxBdl, c, event); err != nil {
			return err
		}
		return nil
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	if err := wt.RenderList(c, event); err != nil {
		return err
	}
	if err := wt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}
func (wt *WorkflowTable) RenderList(component *apistructs.Component, event apistructs.ComponentEvent) error {
	var (
		nodeList []*v1.Node
		nodes    []*v1.Node
		filter   string
	)
	if wt.State.PageNo == 0 {
		wt.State.PageNo = DefaultPageNo
	}
	if wt.State.PageSize == 0 {
		wt.State.PageSize = DefaultPageSize
	}
	pageNo := wt.State.PageNo
	pageSize := wt.State.PageSize
	filter = wt.State.Query["title"].(string)
	//orgID, err := strconv.ParseInt(wt.CtxBdl.Identity.OrgID, 10, 64)
	//clusterNames, err := bdl.Bdl.ListClusters("", uint64(orgID))
	//// Get all nodes by cluster name
	//for _, clusterName := range clusterNames {
	//	nodeReq := &apistructs.{}
	//	nodeReq.ClusterName = clusterName.Name
	//	if nodes, err = bdl.Bdl.ListNodes(nodeReq); err != nil {
	//		return err
	//	}
	//	nodeList = append(nodeList, nodes...)
	//}
	if filter == "" {
		nodes = nodeList
	} else {
		nodes = nodes[:0]
		// Filter by node name or node uid
		for _, node := range nodeList {
			if strings.Contains(node.Name, filter) || strings.Contains(string(node.UID), filter) {
				nodes = append(nodes, node)
			}
		}
	}
	nodes = nodes[(pageNo-1)*pageSize : mathutil.Max((pageNo-1)*pageSize, len(nodes))]
	return wt.setData(nodes)
}

// SetComponentValue transfer WorkflowTable struct to Component
func (wt *WorkflowTable) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(wt.State); err != nil {
		return err
	}
	c.State = state
	c.Operations = wt.Operations
	c.Data["list"] = wt.Data
	return nil
}

//func getOperations(clickableKeys []uint64) map[string]interface{} {
//	return map[string]interface{}{
//		"changePageNo": Operation{
//			Key:    "changePageNo",
//			Reload: true,
//		},
//		"clickRow": Operation{
//			Key:           "clickRow",
//			Reload:        true,
//			FillMeta:      "target",
//			Meta:          nil,
//			ClickableKeys: clickableKeys,
//		},
//	}
//}

func getProps() map[string]interface{} {
	return tableProperties
}
func getTableOperation() map[string]interface{} {
	ops := map[string]Operation{
		"changePageNo": {
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": {
			Key:    "changePageSize",
			Reload: true,
		},
	}
	res := map[string]interface{}{}
	for key, op := range ops {
		res[key] = interface{}(op)
	}
	return res
}

// setData assemble rowItem of table
func (wt *WorkflowTable) setData(nodes []*v1.Node) error {
	var (
		lists []RowItem
		ri    *RowItem
		err   error
	)
	wt.State.Total = len(nodes)
	// todo : return data sorted by column?
	start := (wt.State.PageNo - 1) * wt.State.PageSize
	end := mathutil.Max(wt.State.PageNo*wt.State.PageSize, wt.State.Total)

	for i := start; i < end; i++ {
		if ri, err = wt.getRowItem(nodes[i]); err != nil {
			return err
		}
		lists = append(lists, *ri)
	}
	wt.Data = lists
	return nil
}
func (wt *WorkflowTable) getRowItem(node *v1.Node) (*RowItem, error) {
	var (
		status                  *common.SteveStatus
		distribution, dr, usage *DistributionValue
	)

	nodeLabels := node.GetLabels()

	ri := &RowItem{
		Version: node.Status.NodeInfo.KubeletVersion,
		Role:    getRole(nodeLabels),
		Labels: labels{
			RenderType: "tagsColumn",
			Value:      getPodLabels(node.GetLabels()),
			Operation:  getLabelOperation(string(node.UID)),
		},
		Node: Node{
			RenderType: "linkText",
			Value:      getNodeAddress(node.Status.Addresses),
			Operation:  getNodeOperation(),
			Reload:     false,
		},
		// todo : summarize roles in each pods
		Status: *status,

		// pods total mem / allocate
		Distribution: Distribution{
			RenderType: "bgProgress",
			Value:      *distribution,
		},
		// cmp/customDashboard backend api
		Usage: Distribution{
			RenderType: "bgProgress",
			Value:      *usage,
		},

		DistributionRate: Distribution{
			RenderType: "bgProgress",
			Value:      *dr,
		},
	}
	return ri, nil
}

func getItemStatus(node *v1.Node) (*common.SteveStatus, error) {
	if node == nil {
		return nil, common.NodeNotFoundErr
	}
	ss := &common.SteveStatus{
		RenderType: "textWithBadge",
	}
	status := common.NodeStatusReady
	if node.Spec.Unschedulable {
		status = common.NodeStatusFreeze
	} else {
		for _, cond := range node.Status.Conditions {
			if cond.Status == v1.ConditionTrue && cond.Type == v1.NodeReady {
				status = common.NodeStatusError
				break
			}
		}
	}
	// 0:English 1:ZH
	ss.Status = common.GetNodeStatus(status)[0]
	ss.Value = common.GetNodeStatus(status)[1]
	return ss, nil
}
func (wt *WorkflowTable) getUsageValue(node *v1.Node) (*DistributionValue, error) {
	var (
		resp *pb.QueryWithInfluxFormatResponse
		err  error
	)
	start := time.Now().Nanosecond()
	req := &pb.QueryWithInfluxFormatRequest{
		Start:   strconv.Itoa(start),
		End:     strconv.Itoa(start),
		Filters: nil,
		Options: nil,
		Statement: `SELECT mem_used , mem_total , mem_used_percent FROM status_page 
	WHERE cluster_name::tag=$cluster_name && hostname::tag=$hostname
	ORDER BY TIMESTAMP DESC`,
		Params: map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue(node.ClusterName),
			"$hostname":    structpb.NewStringValue(node.Name),
		},
	}
	if resp, err = metricsServer.QueryWithInfluxFormat(context.Background(), req); err != nil {
		return nil, err
	}
	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		serie := resp.Results[0].Series[0]
		usageDecimal := serie.Rows[0].Values[0].GetNumberValue()
		totalDecimal := serie.Rows[0].Values[1].GetNumberValue()
		usageDecimal, totalDecimal = common.ResetNumberBase(usageDecimal, totalDecimal)
		usageRate := serie.Rows[0].Values[2].GetNumberValue()
		return &DistributionValue{
			Text:    fmt.Sprintf("%.1f/%.1f", usageRate, totalDecimal),
			Percent: int(usageRate),
		}, nil
	}
	return nil, common.ResourceEmptyErr

}
func getRole(labels map[string]string) string {
	res := make([]string, 0)
	for k, _ := range labels {
		if strings.HasPrefix(k, "node-role") {
			splits := strings.Split(k, "\\")
			res = append(res, splits[len(splits)-1])
		}
	}
	return strutil.Join(res, ",", true)
}
func getPodLabels(labels map[string]string) []LabelsValue {
	labelValues := make([]LabelsValue, 0)
	for key, value := range labels {
		lv := LabelsValue{
			Label: fmt.Sprintf("%s=%s", key, value),
			// todo group
			Group: "",
		}
		labelValues = append(labelValues, lv)
	}
	return labelValues
}

func getLabelOperation(rowId string) map[string]Operation {
	return map[string]Operation{
		"add": {
			Key:    "addLabel",
			Reload: false,
			Command: Command{
				Key: "set",
				Command: CommandState{
					Visible:  true,
					FromData: FromData{RecordId: rowId},
				},
				Target: "addLabelModel",
			},
		},
		"delete": {
			Key:      "deleteLabel",
			Reload:   false,
			FillMeta: "label",
			Meta: map[string]string{
				"RecordId": rowId,
			},
		},
	}
}
func getNodeOperation() map[string]Operation {
	return map[string]Operation{
		"click": {Key: "goto", Target: "orgRoot"},
	}
}
func getNodeAddress(addrs []v1.NodeAddress) string {
	for _, addr := range addrs {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
func (wt *WorkflowTable) updateTable(c *apistructs.Component) error {
	var (
		stateValue []byte
		err        error
		state      map[string]interface{}
	)

	if stateValue, err = json.Marshal(c.State); err != nil {
		return err
	}

	if err = json.Unmarshal(stateValue, &state); err != nil {
		return err
	}

	c.State = state
	c.Type = wt.Type

	// export rendered components data
	c.Operations = wt.Operations
	c.Props = getProps()
	return nil
}

func (wt *WorkflowTable) listOperationHandler(bdl protocol.ContextBundle) error {
	var	nodeList []*v1.Node

	return wt.setData(nodeList)
}

// TODO click row will show node detail
func (wt *WorkflowTable) clickRowOperationHandler(bdl protocol.ContextBundle, c *apistructs.Component, event apistructs.ComponentEvent) error {

	return nil
}

func RenderCreator() protocol.CompRender {
	return nil
}

func GetOpsInfo(opsData interface{}) (*Meta, error) {
	if opsData == nil {
		return nil, common.OperationsEmptyErr
	}
	var op Operation
	cont, err := json.Marshal(opsData)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", opsData, err)
		return nil, err
	}
	err = json.Unmarshal(cont, &op)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	meta := op.Meta.(Meta)
	return &meta, nil
}
func (wt *WorkflowTable) RenderChangePageSize(ops interface{}) error {
	meta, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	wt.State.PageNo = 1
	wt.State.PageSize = meta.PageSize
	return nil
}

func (wt *WorkflowTable) RenderChangePageNo(ops interface{}) error {
	meta, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	wt.State.PageNo = meta.PageNo
	wt.Props = getProps()
	return nil
}
