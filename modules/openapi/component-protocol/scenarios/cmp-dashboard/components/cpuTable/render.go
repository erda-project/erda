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

package cpuTable

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/rancher/norman/types/convert"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/api/resource"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

var metricsServer = servicehub.New().Service("metrics-query").(pb.MetricServiceServer)

var tableProperties = map[string]interface{}{
	"rowKey": "id",
	"columns": []Columns{
		{DataIndex: "status", Title: "状态"},
		{DataIndex: "node", Title: "节点"},
		{DataIndex: "role", Title: "角色"},
		{DataIndex: "version", Title: "版本"},
		{DataIndex: "distribuTion", Title: "cpu分配率"},
		{DataIndex: "use", Title: "cpu使用率"},
		{DataIndex: "distribuTionRate", Title: "cpu分配使用率"},
	},
	"bordered":        true,
	"selectable":      true,
	"pageSizeOptions": []string{"10", "20", "50", "100"},
}

func (ct *CpuInfoTable) Import(c *apistructs.Component) error {
	var (
		b   []byte
		err error
	)
	if b, err = json.Marshal(c); err != nil {
		return err
	}
	if err = json.Unmarshal(b, ct); err != nil {
		return err
	}
	return nil
}

func (ct *CpuInfoTable) Export(c *apistructs.Component, gs *apistructs.GlobalStateData) error {
	// set components data
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, ct); err != nil {
		return err
	}
	return nil
}

// GenComponentState 获取state
// todo: move to common ?
func (ct *CpuInfoTable) GenComponentState(c *apistructs.Component) error {
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
	ct.State = state
	return nil
}
func (ct *CpuInfoTable) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	ct.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err := ct.GenComponentState(c)
	if err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
		if err := ct.listOperationHandler(ct.CtxBdl); err != nil {
			return err
		}
	case apistructs.CMPDashboardChangePageSizeOperationKey:
		if err := ct.RenderChangePageSize(event.OperationData); err != nil {
			return err
		}
	case apistructs.CMPDashboardChangePageNoOperationKey:
		if err := ct.RenderChangePageNo(event.OperationData); err != nil {
			return err
		}
	case apistructs.RenderingOperation:
		// IsFirstFilter delivered from filer component
		if ct.State.IsFirstFilter {
			ct.State.PageNo = 1
			ct.State.IsFirstFilter = false
		}
	case apistructs.ExecuteClickRowNoOperationKey:
		if err := ct.clickRowOperationHandler(ct.CtxBdl, c, event); err != nil {
			return err
		}
		return nil
	case apistructs.CMPDashboardSortByColumnOperationKey:
		ct.State.PageNo = 1
		ct.State.IsFirstFilter = false
		return nil
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	if err := ct.RenderList(c, event); err != nil {
		return err
	}
	if err := ct.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}
func (ct *CpuInfoTable) RenderList(component *apistructs.Component, event apistructs.ComponentEvent) error {
	var (
		nodeList     []*v1.Node
		nodes        []*v1.Node
		err          error
		filter       string
		sortColumn   string
		clusterNames []apistructs.ClusterInfo
	)
	if ct.State.PageNo == 0 {
		ct.State.PageNo = DefaultPageNo
	}
	if ct.State.PageSize == 0 {
		ct.State.PageSize = DefaultPageSize
	}
	pageNo := ct.State.PageNo
	pageSize := ct.State.PageSize
	filter = ct.State.Query["title"].(string)
	sortColumn = ct.State.SortColumnName
	orgID, err := strconv.ParseInt(ct.CtxBdl.Identity.OrgID, 10, 64)
	if ct.State.ClusterName != "" {
		clusterNames = append([]apistructs.ClusterInfo{}, apistructs.ClusterInfo{Name: ct.State.ClusterName})
	} else {
		clusterNames, err = bdl.Bdl.ListClusters("", uint64(orgID))
		if err != nil {
			return err
		}
	}
	// Get all nodes by cluster name
	for _, clusterName := range clusterNames {
		nodeReq := &apistructs.K8SResourceRequest{}
		nodeReq.ClusterName = clusterName.Name
		if nodes, err = bdl.Bdl.ListNodes(nodeReq); err != nil {
			return err
		}
		nodeList = append(nodeList, nodes...)
	}
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
	if sortColumn != "" {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Name < nodes[j].Name
		})
	}
	nodes = nodes[(pageNo-1)*pageSize : mathutil.Max((pageNo-1)*pageSize, len(nodes))]
	return ct.setData(nodes)
}

// SetComponentValue mapping CpuInfoTable properties to Component
func (ct *CpuInfoTable) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(ct.State); err != nil {
		return err
	}
	c.State = state
	c.Operations = ct.Operations
	c.Data["list"] = ct.Data
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

func getItemOperation(id string) map[string]interface{} {
	ops := map[string]Operation{
		"add": {
			Key:    "addLabel",
			Reload: false,
			Command: Command{
				Key:     "goto",
				Command: CommandState{true, FromData{RecordId: id}},
				Target:  "orgRoot",
			},
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
func (ct *CpuInfoTable) setData(nodes []apistructs.SteveResource) error {
	var (
		lists []RowItem
		ri    *RowItem
		err   error
	)
	ct.State.Total = len(nodes)
	// todo : return data sorted by column?
	start := (ct.State.PageNo - 1) * ct.State.PageSize
	end := mathutil.Max(ct.State.PageNo*ct.State.PageSize, ct.State.Total)

	for i := start; i < end; i++ {
		if ri, err = ct.getRowItem(nodes[i]); err != nil {
			return err
		}
		lists = append(lists, *ri)
	}
	ct.Data = lists
	return nil
}
func (ct *CpuInfoTable) getDistributionValue(node apistructs.SteveResource) (*DistributionValue, error) {
	var (
		pods     []apistructs.SteveResource
		err      error
		cpuValue resource.Quantity
	)
	req := &apistructs.SteveRequest{
		ClusterName:
		node.ClusterName,
			Namespace:     node.Namespace,
		LabelSelector: []string{fmt.Sprintf("=%s", node.Name)},
	}
	if pods, err = ct.CtxBdl.Bdl.ListPods(req); err != nil {
		return nil, err
	}
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			cpuValue.Add(container.Resources.Requests.Cpu().DeepCopy())
		}
	}
	allocValue := node.Status.Allocatable.Cpu().Value()
	allocDecimal := float64(allocValue)
	usageDecimal := float64(cpuValue.Value())
	return &DistributionValue{
		Text:    fmt.Sprintf("%.1f/%.1f", usageDecimal, allocDecimal),
		Percent: int(usageDecimal * 100 / allocDecimal),
	}, nil
}
func (ct *CpuInfoTable) getDistributionRate(node *v1.Node) (*DistributionValue, error) {
	cpuAllocatable := node.Status.Allocatable.Cpu().Value()
	cpuCapcity := node.Status.Capacity.Cpu().Value()
	common.GetInt64Len(cpuAllocatable)
	return &DistributionValue{
		Text:    fmt.Sprintf("%d/%d", cpuAllocatable, cpuCapcity),
		Percent: common.GetPercent(float64(cpuAllocatable), float64(cpuCapcity)),
	}, nil
}
func (ct *CpuInfoTable) getRowItem(c apistructs.SteveResource) (*RowItem, error) {
	var (
		err                     error
		status                  *common.SteveStatus
		distribution, dr, usage *DistributionValue
	)

	nodeLabels := c.Metadata.Labels
	status = getItemStatus(c)
	if distribution, err = ct.getDistributionValue(c); err != nil {
		return nil, err
	}
	if usage, err = ct.getUsageValue(c); err != nil {
		return nil, err
	}
	if dr, err = ct.getDistributionRate(c); err != nil {
		return nil, err
	}
	ri := &RowItem{
		ID:      node.Name,
		Version: node.Status.NodeInfo.KubeletVersion,
		Role:    getRole(nodeLabels),
		Labels: Labels{
			RenderType: "tagsColumn",
			Value:      getPodLabels(node.GetLabels()),
			Operation:  getLabelOperation(string(node.UID)),
		},
		Node: Node{
			RenderType: "linkText",
			Value:      getNodeAddress(node.Status.Addresses),
			Operation:  getNodeOperation(),
			reload:     false,
		},
		Status: *status,
		Distribution: Distribution{
			RenderType: "bgProgress",
			Value:      *distribution,
		},
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
func (ct *CpuInfoTable) getUsageValue(node *v1.Node) (*DistributionValue, error) {
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
		Statement: `SELECT cpu_cores_usage ,n_cpus ,cpu_usage_active FROM status_page 
	WHERE cluster_name::tag=$cluster_name && hostname::tag=$hostname
	ORDER BY TIMESTAMP DESC`,
		Params: map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue(node.ClusterName),
			"hostname":     structpb.NewStringValue(node.Name),
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
	return nil, common.NodeStatusEmptyErr

}
func getItemStatus(node apistructs.SteveResource) *common.SteveStatus {
	ss := &common.SteveStatus{
		RenderType: "textWithBadge",
	}
	status := common.NodeStatusReady
	spec := convert.ToMapInterface(node.Spec)
	if unscheduled, ok := spec["unschedulable"]; ok && unscheduled.(bool) {
		status = common.NodeStatusFreeze
	} else {
		nodeStatus := convert.ToMapInterface(node.Status)
		if conditions, ok := nodeStatus["conditions"];ok{
			for _, cond := range conditions.([]interface{}){
				condItem := convert.ToMapInterface(cond)
			if condItem["condItem"] == v1.ConditionTrue && condItem["Type"] == v1.NodeReady{
				status = common.NodeStatusError
			break
		}
		}
		}
	}
	// 0:English 1:ZH
	ss.Status = common.GetNodeStatus(status)[0]
	ss.Value = common.GetNodeStatus(status)[1]
	return ss
}

//func getDistributionValue(node *v1.Node) (*DistributionValue, error) {
//	if node == nil {
//		return nil, common.NodeNotFoundErr
//	}
//	return &DistributionValue{
//		Text:    "",
//		Percent: 0,
//	}, nil
//}
//func getDistributionRate(node *v1.Node) DistributionValue {
//	cpuAllocatable := node.Status.Allocatable.Cpu().Value()
//	cpuCapacity := node.Status.Capacity.Cpu().Value()
//	baseNum := math.Pow(10, float64(mathutil.Min(common.GetInt64Len(cpuCapacity), common.GetInt64Len(cpuAllocatable))))
//	capDecimal := float64(cpuCapacity) / baseNum
//	allocDecimal := float64(cpuAllocatable) / baseNum
//
//	return DistributionValue{
//		Text:    fmt.Sprintf("%.1f/%.1f", allocDecimal, capDecimal),
//		Percent: common.GetPercent(allocDecimal, capDecimal),
//	}
//}
//func getUsageValue(node *v1.Node) DistributionValue {
//	capacity := node.Status.Capacity.Cpu().Value()
//	alloc := node.Status.Allocatable.Cpu().Value()
//	return DistributionValue{
//		Text:    "",
//		Percent: int(alloc / capacity),
//	}
//}
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

func (ct *CpuInfoTable) listOperationHandler(bdl protocol.ContextBundle) error {
	var (
		nodeList []apistructs.SteveResource
		nodes    *apistructs.SteveCollection
		err      error
	)
	orgID, err := strconv.ParseInt(ct.CtxBdl.Identity.OrgID, 10, 64)
	clusterNames, err := bdl.Bdl.ListClusters("", uint64(orgID))

	for _, clusterName := range clusterNames {
		nodeReq := &apistructs.SteveRequest{}
		nodeReq.ClusterName = clusterName.Name
		if nodes, err = bdl.Bdl.ListSteveResource(nodeReq); err != nil {
			return err
		}
		nodeList = append(nodeList, nodes.Data...)
	}
	return ct.setData(nodeList)
}

// TODO click row will show node detail
func (ct *CpuInfoTable) clickRowOperationHandler(bdl protocol.ContextBundle, c *apistructs.Component, event apistructs.ComponentEvent) error {

	return nil
}

func RenderCreator() protocol.CompRender {
	return &CpuInfoTable{
		Type:       "Table",
		Props:      getProps(),
		Operations: getTableOperation(),
		State:      common.State{},
	}
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
func (ct *CpuInfoTable) RenderChangePageSize(ops interface{}) error {
	meta, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	ct.State.PageNo = 1
	ct.State.PageSize = meta.PageSize
	return nil
}

func (ct *CpuInfoTable) RenderChangePageNo(ops interface{}) error {
	meta, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	ct.State.PageNo = meta.PageNo
	ct.Props = getProps()
	return nil
}

func (ct *CpuInfoTable) RenderSortColumn(ops interface{}) error {
	meta, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	ct.State.SortColumnName = meta.SortColumn
	ct.Props = getProps()
	return nil
}
