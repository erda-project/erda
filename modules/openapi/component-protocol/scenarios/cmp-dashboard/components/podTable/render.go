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
	"github.com/cznic/mathutil"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/resourceinfo"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

var tableProperties = map[string]interface{}{
	"rowKey": "name",
	"columns": []Columns{
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

	switch event.Operation {
	case apistructs.InitializeOperation:
		if err := pt.listOperationHandler(pt.CtxBdl); err != nil {
			return err
		}
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
	case apistructs.ExecuteClickRowNoOperationKey:
		if err := pt.clickRowOperationHandler(pt.CtxBdl, c, event); err != nil {
			return err
		}
		return nil
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
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
		podList []*v1.Pod
		pods    []*v1.Pod
		err      error
		filter   string
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
	orgID, err := strconv.ParseInt(pt.CtxBdl.Identity.OrgID, 10, 64)
	clusterNames, err := bdl.Bdl.ListClusters("", uint64(orgID))
	// Get all pods by cluster name
	for _, clusterName := range clusterNames {
		nodeReq := &apistructs.K8SResourceRequest{}
		nodeReq.ClusterName = clusterName.Name
		if pods, err = bdl.Bdl.ListPods(nodeReq); err != nil {
			return err
		}
		podList = append(podList, pods...)
	}
	if filter == "" {
		pods = podList
	} else {
		pods = pods[:0]
		// Filter by pod name or pod uid
		for _, pod := range podList {
			if strings.Contains(pod.Name, filter) || strings.Contains(string(pod.UID), filter) {
				pods = append(pods, pod)
			}
		}
	}
	pods = pods[(pageNo-1)*pageSize : mathutil.Max((pageNo-1)*pageSize, len(pods))]
	return pt.setData(pods)
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
func (pt *PodInfoTable) setData(pods []*v1.Pod) error {
	var (
		lists []RowItem
		ri    *RowItem
		err   error
	)
	pt.State.Total = len(pods)
	// todo : return data sorted by column?
	start := (pt.State.PageNo - 1) * pt.State.PageSize
	end := mathutil.Max(pt.State.PageNo*pt.State.PageSize, pt.State.Total)

	for i := start; i < end; i++ {
		if ri, err = pt.getRowItem(pods[i]); err != nil {
			return err
		}
		lists = append(lists, *ri)
	}
	pt.Data = lists
	return nil
}
func (pt *PodInfoTable) getRowItem(pod *v1.Pod) (*RowItem, error) {
	var (
		err    error
		status *common.SteveStatus
		CpuUsage, CpuRate, MemUsage, MemRate string
	)

	if status, err = getItemStatus(pod); err != nil {
		return nil, err
	}
	CpuUsage, CpuRate, MemUsage, MemRate,err= getResourceDistributionAndUsage(pod)
	if err != nil{
		return nil,err
	}
	status ,err= getItemStatus(pod)
	if err != nil{
		return nil,err
	}
	ri := &RowItem{
		ID:              pod.Name,
		Status:          *status,
		Namespace:       pod.Namespace,
		CpuUsage:        CpuUsage,
		CpuRate:         CpuRate,
		MemUsage:        MemUsage,
		MemRate:         MemRate,
		RestartTimes:    getRestartTimes(pod),
		ReadyContainers: getReadyContainers(pod),
		PodIp:           pod.Status.PodIP,
		Workload:        getWorkload(pod),
	}
	return ri, nil
}

// getResourceDistributionAndUsage returns CpuUsage CpuRate MemUsage MemRate
func getResourceDistributionAndUsage(pod *v1.Pod) (string, string, string, string,error) {
	if pod == nil {
		return "","","","",common.PodNotFoundErr
	}
	req, limits := resourceinfo.PodRequestsAndLimits(pod)
	cpuUsage := req.Cpu().String()
	cpuRate := fmt.Sprintf("%.1f", float64(req.Cpu().Value()*100)/float64(limits.Cpu().Value()))
	memUsage := req.Memory().String()
	memRate := fmt.Sprintf("%.1f", float64(req.Memory().Value()*100)/float64(limits.Memory().Value()))
	return cpuUsage, cpuRate, memUsage, memRate,nil
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
	statuses:=common.GetPodStatus(common.SteveStatusEnum(pod.Status.Phase))
	ss.Status =statuses[0]
	ss.Value = statuses[1]
	return ss, nil
}

func getRole(labels map[string]string) string {
	res := make([]string, 0)
	for k, _ := range labels {
		if strings.HasPrefix(k, "pod-role") {
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

func (pt *PodInfoTable) listOperationHandler(bdl protocol.ContextBundle) error {
	var (
		podList []*v1.Pod
		pods    []*v1.Pod
		err     error
	)
	orgID, err := strconv.ParseInt(pt.CtxBdl.Identity.OrgID, 10, 64)
	clusterNames, err := bdl.Bdl.ListClusters("", uint64(orgID))

	for _, clusterName := range clusterNames {
		podReq := &apistructs.K8SResourceRequest{}
		podReq.ClusterName = clusterName.Name
		if pods, err = bdl.Bdl.ListPods(podReq); err != nil {
			return err
		}
		podList = append(podList, pods...)
	}
	return pt.setData(podList)
}

// TODO click row will show pod detail
func (pt *PodInfoTable) clickRowOperationHandler(bdl protocol.ContextBundle, c *apistructs.Component, event apistructs.ComponentEvent) error {

	return nil
}

func RenderCreator() protocol.CompRender {
	return &PodInfoTable{
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
