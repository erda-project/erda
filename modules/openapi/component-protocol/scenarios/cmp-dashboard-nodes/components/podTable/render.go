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

package podTable

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/components/tab"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/resourceinfo"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

var tableProperties = map[string]interface{}{
	"rowKey": "name",
	"columns": []common.Columns{
		{DataIndex: "status", Title: "状态"},
		{DataIndex: "pod", Title: "名称"},
		{DataIndex: "role", Title: "命名空间"},
		{DataIndex: "distribuTion", Title: "cpu分配量"},
		{DataIndex: "use", Title: "cpu水位(使用量/限制量)"},
		{DataIndex: "distribuTionRate", Title: "内存分配量"},
		{DataIndex: "distribuTionRate", Title: "内存水位(使用量/限制量)"},
		{DataIndex: "distribuTionRate", Title: "容器就绪"},
		{DataIndex: "distribuTionRate", Title: "重启次数"},
		{DataIndex: "distribuTionRate", Title: "IP"},
		{DataIndex: "distribuTionRate", Title: "节点"},
		{DataIndex: "distribuTionRate", Title: "存活时间"},
		{DataIndex: "distribuTionRate", Title: "工作负载"},
	},
	"bordered":        true,
	"selectable":      true,
	"pageSizeOptions": []string{"10", "20", "50", "100"},
}
var metricsServer = servicehub.New().Service("metrics-query").(pb.MetricServiceServer)

// GenComponentState 获取state
func (pt *PodInfoTable) GenComponentState(c *apistructs.Component) error {
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
	pt.State = state
	return nil
}
func (pt *PodInfoTable) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	pt.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err := pt.GenComponentState(c)
	if err != nil {
		return err
	}
	if event.Operation != apistructs.InitializeOperation {
		if c.State["activeKey"] != tab.POD_TAB {
			return nil
		}
		switch event.Operation {

		case apistructs.CMPDashboardChangePageSizeOperationKey:
			if err := pt.RenderChangePageSize(event.OperationData); err != nil {
				return err
			}
		case apistructs.CMPDashboardChangePageNoOperationKey:
			if err := pt.RenderChangePageNo(event.OperationData); err != nil {
				return err
			}
		case apistructs.RenderingOperation:
			// IsFirstFilter delivered from filer component
			if pt.State.IsFirstFilter {
				pt.State.PageNo = 1
				pt.State.IsFirstFilter = false
			}
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	}
	if err := pt.RenderList(c, event); err != nil {
		return err
	}
	if err := pt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}
func (pt *PodInfoTable) RenderList(component *apistructs.Component, event apistructs.ComponentEvent) error {
	var (
		nodeList     []apistructs.SteveResource
		nodes        []apistructs.SteveResource
		resp         *apistructs.SteveCollection
		err          error
		filter       string
		sortColumn   string
		orgID        int64
		asc          bool
		clusterNames []apistructs.ClusterInfo
	)
	if pt.State.PageNo == 0 {
		pt.State.PageNo = DefaultPageNo
	}
	if pt.State.PageSize == 0 {
		pt.State.PageSize = DefaultPageSize
	}
	pageNo := pt.State.PageNo
	pageSize := pt.State.PageSize
	filter = pt.State.Query["title"].(string)
	sortColumn = pt.State.SortColumnName
	asc = pt.State.Asc

	if pt.State.ClusterName != "" {
		clusterNames = append([]apistructs.ClusterInfo{}, apistructs.ClusterInfo{Name: pt.State.ClusterName})
	} else {
		clusterNames, err = bdl.Bdl.ListClusters("", uint64(orgID))
		if err != nil {
			return err
		}
	}
	// Get all nodes by cluster name
	for _, clusterName := range clusterNames {
		nodeReq := &apistructs.SteveRequest{}
		nodeReq.Name = clusterName.Name
		nodeReq.ClusterName = clusterName.Name
		resp, err = bdl.Bdl.ListSteveResource(nodeReq)
		if err != nil {
			return err
		}
		nodeList = append(nodeList, resp.Data...)
	}
	if filter == "" {
		nodes = nodeList
	} else {
		// Filter by node name or node uid
		for _, node := range nodeList {
			if strings.Contains(node.Metadata.Name, filter) || strings.Contains(node.ID, filter) {
				nodes = append(nodes, node)
			}
		}
	}
	if err = pt.SetData(nodes, v1.ResourceMemory); err != nil {
		return err
	}

	if sortColumn != "" {
		refCol := reflect.ValueOf(common.RowItem{}).FieldByName(sortColumn)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			common.SortByString([]interface{}{pt.Data}, sortColumn, asc)
		case reflect.TypeOf(common.Node{}):
			common.SortByNode([]interface{}{pt.Data}, sortColumn, asc)
		case reflect.TypeOf(common.Distribution{}):
			common.SortByDistribution([]interface{}{pt.Data}, sortColumn, asc)
		default:
			logrus.Errorf("sort by column %s not found", sortColumn)
			return common.TypeNotAvailableErr
		}
	}
	nodes = nodes[(pageNo-1)*pageSize : mathutil.Max((pageNo-1)*pageSize, len(nodes))]
	component.Data["list"] = nodes
	return nil
}

// SetComponentValue transfer CpuInfoTable struct to Component
func (pt *PodInfoTable) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(pt.State); err != nil {
		return err
	}
	c.State = state
	c.Operations = pt.Operations
	c.Data["list"] = pt.Data
	return nil
}

func getProps() map[string]interface{} {
	return tableProperties
}
func getTableOperation() map[string]interface{} {
	ops := map[string]common.Operation{
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

// SetData assemble rowItem of table

func (pt *PodInfoTable) SetData(pods []apistructs.SteveResource, resName v1.ResourceName) error {
	var (
		lists []RowItem
		ri    *RowItem
		err   error
	)

	pt.State.Total = len(pods)

	start := (pt.State.PageNo - 1) * pt.State.PageSize
	end := mathutil.Max(pt.State.PageNo*pt.State.PageSize, pt.State.Total)

	for i := start; i < end; i++ {
		if ri, err = pt.GetRowItem(pods[i]); err != nil {
			return err
		}
		lists = append(lists, *ri)
	}
	pt.Data = lists
	return nil
}
func (pt *PodInfoTable) GetRowItem(pod apistructs.SteveResource) (*RowItem, error) {
	var (
		err                                  error
		status                               *common.SteveStatus
		k8spod                               = v1.Pod{}
		CpuUsage, CpuRate, MemUsage, MemRate string
	)
	err = common.Transfer(pod, k8spod)
	if err != nil {
		return nil, err
	}
	if status, err = getItemStatus(&k8spod); err != nil {
		return nil, err
	}
	CpuUsage, CpuRate, MemUsage, MemRate, err = getResourceDistributionAndUsage(&k8spod)
	if err != nil {
		return nil, err
	}
	status, err = getItemStatus(&k8spod)
	if err != nil {
		return nil, err
	}
	ri := &RowItem{
		ID:              k8spod.Name,
		Status:          *status,
		Namespace:       k8spod.Namespace,
		CpuUsage:        CpuUsage,
		CpuRate:         CpuRate,
		MemUsage:        MemUsage,
		MemRate:         MemRate,
		RestartTimes:    getRestartTimes(&k8spod),
		ReadyContainers: getReadyContainers(&k8spod),
		PodIp:           k8spod.Status.PodIP,
		Workload:        getWorkload(&k8spod),
	}
	return ri, nil
}

// getResourceDistributionAndUsage returns CpuUsage CpuRate MemUsage MemRate
func getResourceDistributionAndUsage(pod *v1.Pod) (string, string, string, string, error) {
	if pod == nil {
		return "", "", "", "", common.PodNotFoundErr
	}
	req, limits := resourceinfo.PodRequestsAndLimits(pod)
	cpuUsage := req.Cpu().String()
	cpuRate := fmt.Sprintf("%.1f", float64(req.Cpu().Value()*100)/float64(limits.Cpu().Value()))
	memUsage := req.Memory().String()
	memRate := fmt.Sprintf("%.1f", float64(req.Memory().Value()*100)/float64(limits.Memory().Value()))
	return cpuUsage, cpuRate, memUsage, memRate, nil
}

func getWorkload(pod *v1.Pod) string {
	for key, label := range pod.Labels {
		if strings.Contains(key, "workload") {
			return label
		}
	}
	return ""
}

func getReadyContainers(pod *v1.Pod) int {
	var cnt = 0
	for _, container := range pod.Status.ContainerStatuses {
		if container.Ready {
			cnt++
		}
	}
	return cnt
}
func getRestartTimes(pod *v1.Pod) int {
	var max = 0
	for _, container := range pod.Status.ContainerStatuses {
		max = mathutil.Max(max, int(container.RestartCount))
	}
	return max
}

func getItemStatus(pod *v1.Pod) (*common.SteveStatus, error) {
	if pod == nil {
		return nil, common.NodeNotFoundErr
	}
	ss := &common.SteveStatus{
		RenderType: "textWithBadge",
	}

	// 0:English 1:ZH
	statuses := common.GetPodStatus(common.SteveStatusEnum(pod.Status.Phase))
	ss.Status = statuses[0]
	ss.Value = statuses[1]
	return ss, nil
}

func getRole(labels map[string]string) string {
	res := make([]string, 0)
	for k := range labels {
		if strings.HasPrefix(k, "pod-role") {
			splits := strings.Split(k, "\\")
			res = append(res, splits[len(splits)-1])
		}
	}
	return strutil.Join(res, ",", true)
}
func getPodLabels(labels map[string]string) []common.LabelsValue {
	labelValues := make([]common.LabelsValue, 0)
	for key, value := range labels {
		lv := common.LabelsValue{
			Label: fmt.Sprintf("%s=%s", key, value),
			// todo group
			Group: "",
		}
		labelValues = append(labelValues, lv)
	}
	return labelValues
}

func getLabelOperation(rowId string) map[string]common.Operation {
	return map[string]common.Operation{
		"add": {
			Key:    "addLabel",
			Reload: false,
			Command: common.Command{
				Key: "set",
				Command: common.CommandState{
					Visible:  true,
					FromData: common.FromData{RecordId: rowId},
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

func getNodeOperation() map[string]common.Operation {
	return map[string]common.Operation{
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

func (pt *PodInfoTable) updateTable(c *apistructs.Component) error {
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
	c.Type = pt.Type

	// export rendered components data
	c.Operations = pt.Operations
	c.Props = getProps()
	return nil
}

func RenderCreator() protocol.CompRender {
	pi := PodInfoTable{}
	pi.Type = "Table"
	pi.Props = getProps()
	pi.Operations = getTableOperation()
	pi.State = common.State{}
	return &pi
}

func GetOpsInfo(opsData interface{}) (*common.Meta, error) {
	if opsData == nil {
		return nil, common.OperationsEmptyErr
	}
	var op common.Operation
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
	meta := op.Meta.(common.Meta)
	return &meta, nil
}
func (pt *PodInfoTable) RenderChangePageSize(ops interface{}) error {
	meta, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	pt.State.PageNo = 1
	pt.State.PageSize = meta.PageSize
	return nil
}

func (pt *PodInfoTable) RenderChangePageNo(ops interface{}) error {
	meta, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	pt.State.PageNo = meta.PageNo
	pt.Props = getProps()
	return nil
}
