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

package table

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/resourceinfo"
	"modernc.org/mathutil"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/metrics"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-pods/common"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

var (
	ops = map[string]Operation{
		"changePageNo": {
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": {
			Key:    "changePageSize",
			Reload: true,
		},
		"changeSorter": {
			Key:    "changePageSize",
			Reload: true,
		},
	}
)

type Table struct {
	TableInterface
	Ctx        context.Context
	Metric     *metrics.Metrics
	CtxBdl     protocol.ContextBundle
	Type       string                 `json:"type"`
	Props      map[string]interface{} `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
	Data       []RowItem              `json:"data"`
}

func (ft *Table) Init(ctx servicehub.Context) error { return nil }

func (ft *Table) Run(ctx context.Context) error {
	return nil
}

type TableInterface interface {
	SetData(resources []apistructs.SteveResource) error
}

type Columns struct {
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
	Width     int    `json:"width,omitempty"`
	Sortable  bool   `json:"sorter"`
}

type Meta struct {
	Id         int    `json:"id,omitempty"`
	PageSize   int    `json:"pageSize,omitempty"`
	PageNo     int    `json:"pageNo,omitempty"`
	SortColumn string `json:"sorter"`
}

type RowItem struct {
	ID        string             `json:"id"`
	Status    common.SteveStatus `json:"status"`
	Pod       Pod                `json:"pod"`
	Namespace string             `json:"namespace"`
	Version   string             `json:"version"`
	//
	Distribution Distribution           `json:"distribution"`
	Usage        string                 `json:"use"`
	Labels       Labels                 `json:"Labels"`
	Operations   map[string]interface{} `json:"operations"`
	Ready        string                 `json:"ready"`
}

type Operation struct {
	Key           string      `json:"key"`
	Reload        bool        `json:"reload"`
	FillMeta      string      `json:"fillMeta,omitempty"`
	Target        string      `json:"target,omitempty"`
	Meta          interface{} `json:"meta,omitempty"`
	ClickableKeys interface{} `json:"clickableKeys,omitempty"`
	Command       Command     `json:"command,omitempty"`
}

type Command struct {
	Key     string       `json:"key"`
	Command CommandState `json:"command"`
	Target  string       `json:"target"`
}

type CommandState struct {
	Visible  bool     `json:"visible"`
	FromData FromData `json:"from_data"`
}

type FromData struct {
	RecordId string `json:"record_id"`
}

type Pod struct {
	RenderType string               `json:"render_type"`
	Value      string               `json:"value"`
	Operation  map[string]Operation `json:"operation"`
	Reload     bool                 `json:"reload"`
}

type Labels struct {
	RenderType string               `json:"render_type"`
	Value      []LabelsValue        `json:"value"`
	Operation  map[string]Operation `json:"operation"`
}

type Distribution struct {
	RenderType string  `json:"render_type"`
	Text       string  `json:"text"`
	Percent    float64 `json:"percent"`
	Status     common.UsageStatusEnum
}

type DistributionValue struct {
}

type LabelsValue struct {
	Label string `json:"label"`
	Group string `json:"group"`
}

// GetResourceDistributionAndUsage returns CpuUsage CpuRate MemUsage MemRate
func (ft *Table) GetResourceDistributionAndUsage(pod *v1.Pod) (string, *Distribution, string, *Distribution, error) {
	if pod == nil {
		return "", nil, "", nil, common.PodNotFoundErr
	}
	req, limits := resourceinfo.PodRequestsAndLimits(pod)
	cpuUsage := req.Cpu().String()
	c := float64(req.Cpu().Value()*100) / float64(limits.Cpu().Value())
	cpuRate := &Distribution{
		RenderType: "progress",
		Text:       fmt.Sprintf("%.1f", c),
		Percent:    c,
		Status:     ft.getDistributionStatus(c),
	}
	memUsage := req.Memory().String()
	m := float64(req.Cpu().Value()*100) / float64(limits.Cpu().Value())
	memRate := &Distribution{
		RenderType: "progress",
		Text:       fmt.Sprintf("%.1f", m),
		Percent:    m,
		Status:     ft.getDistributionStatus(m),
	}
	return cpuUsage, cpuRate, memUsage, memRate, nil
}

func (ft *Table) getDistributionStatus(rate float64) common.UsageStatusEnum {
	switch {
	case rate > 70:
		return common.ResourceSuccess
	case rate > 50:
		return common.ResourceNormal
	case rate > 30:
		return common.ResourceWarning
	case rate > 0:
		return common.ResourceDanger
	default:
		return common.ResourceError
	}
}

func (ft *Table) GetItemStatus(node *v1.Node) (*common.SteveStatus, error) {
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

// SetComponentValue mapping CpuInfoTable properties to Component
func (ft *Table) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(ft.State); err != nil {
		return err
	}
	c.State = state
	return nil
}

// GetOpsInfo return request meta
func (ft *Table) GetOpsInfo(opsData interface{}) (*Meta, error) {
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

func (ft *Table) RenderList(component *apistructs.Component, event apistructs.ComponentEvent) error {
	var (
		list, k8sResources                                     []apistructs.SteveResource
		resp                                                   *apistructs.SteveCollection
		err                                                    error
		namespacesFilter, statusFilter, queryText, clusterName string
		sortColumn                                             SorterData
	)
	if ft.State.PageNo == 0 {
		ft.State.PageNo = DefaultPageNo
	}
	if ft.State.PageSize == 0 {
		ft.State.PageSize = DefaultPageSize
	}

	pageNo := ft.State.PageNo
	pageSize := ft.State.PageSize
	namespacesFilter = ft.State.Query["namespaces"].(string)
	statusFilter = ft.State.Query["status"].(string)
	queryText = ft.State.Query["q"].(string)
	sortColumn = ft.State.SortColumn
	clusterName = ft.CtxBdl.InParams["clusterName"].(string)

	// Get all pods by cluster name
	req := &apistructs.SteveRequest{}
	req.ClusterName = clusterName
	for _, t := range common.ResourcesTypes {
		req.Type = t
		resp, err = bdl.Bdl.ListSteveResource(req)
		if err != nil {
			return err
		}
		list = append(list, resp.Data...)
	}

	// Filter by node name or node uid
	for _, res := range list {
		if err != nil {
			return err
		}
		res.Metadata.
		if strings.Contains(res.Metadata.Name, queryText) && namespacesFilter == res.Metadata.Namespace && string() == statusFilter {
			k8sResources = append(k8sResources, k8sPod)
		}
	}
	if err = ft.SetData(k8sPods); err != nil {
		return err
	}

	if sortColumn.Field != "" {
		refCol := reflect.ValueOf(RowItem{}).FieldByName(sortColumn.Field)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			common.SortByString([]interface{}{ft.Data}, sortColumn.Field, sortColumn.Order)
		case reflect.TypeOf(Pod{}):
			common.SortByNode([]interface{}{ft.Data}, sortColumn.Field, sortColumn.Order)
		default:
			logrus.Errorf("sort by column %s not found", sortColumn)
			return common.TypeNotAvailableErr
		}
	}
	k8sPods = k8sPods[(pageNo-1)*pageSize : mathutil.Max((pageNo-1)*pageSize, len(ft.Data))]
	component.Data["list"] = k8sPods
	return nil
}

// SetData assemble rowItem of table
func (ft *Table) SetData(pods []v1.Pod) error {
	var (
		lists []RowItem
		ri    *RowItem
		err   error
	)

	ft.State.Total = len(pods)

	start := (ft.State.PageNo - 1) * ft.State.PageSize
	end := mathutil.Max(ft.State.PageNo*ft.State.PageSize, ft.State.Total)

	for i := start; i < end; i++ {
		if ri, err = ft.GetRowItem(pods[i]); err != nil {
			return err
		}
		lists = append(lists, *ri)
	}
	ft.Data = lists
	return nil
}

func (ft *Table) RenderChangePageSize(ops interface{}) error {
	meta, err := ft.GetOpsInfo(ops)
	if err != nil {
		return err
	}
	ft.State.PageNo = 1
	ft.State.PageSize = meta.PageSize
	return nil
}

func (ft *Table) RenderChangePageNo(ops interface{}) error {
	meta, err := ft.GetOpsInfo(ops)
	if err != nil {
		return err
	}
	ft.State.PageNo = meta.PageNo
	return nil
}
func (ft *Table) GetPodAddress(addrs []v1.PodIP) string {
	ips := make([]string, len(addrs))
	for i, addr := range addrs {
		ips[i] = addr.IP
	}
	return strings.Join(ips, ",")
}
func (ft *Table) GetRowItem(k8sPod v1.Pod) (*RowItem, error) {
	var (
		err        error
		itemStatus *common.SteveStatus
		cpuUsage   string
		cpuRate    *Distribution
	)

	if itemStatus, err = getItemStatus(&k8sPod); err != nil {
		return nil, err
	}
	cpuUsage, cpuRate, _, _, err = ft.GetResourceDistributionAndUsage(&k8sPod)
	if err != nil {
		return nil, err
	}
	ri := &RowItem{
		ID:     k8sPod.Name,
		Status: *itemStatus,
		Pod: Pod{
			RenderType: "linkText",
			Value:      ft.GetPodAddress(k8sPod.Status.PodIPs),
			Operation:  ft.GetPodOperation(),
			Reload:     false,
		},
		Namespace:    k8sPod.Namespace,
		Usage:        cpuUsage,
		Distribution: *cpuRate,
		Version:      k8sPod.ResourceVersion,
		// todo
		Labels:     GetPodLabels(k8sPod.Labels),
		Operations: GetTableOperation(),
		Ready:      GetReady(&k8sPod),
	}
	return ri, nil
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

func GetPodLabels(labels map[string]string) Labels {
	labelValues := make([]LabelsValue, 0)
	for key, value := range labels {
		lv := LabelsValue{
			Label: fmt.Sprintf("%s=%s", key, value),
			// todo group
			Group: "",
		}
		labelValues = append(labelValues, lv)
	}
	return Labels{
		RenderType: "",
		Value:      labelValues,
		Operation:  nil,
	}
}
func (ft *Table) GetPodLabels(labels map[string]string) []LabelsValue {
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

func GetReady(pod *v1.Pod) string {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == v1.PodReady && cond.Status == v1.ConditionTrue {
			return "1/1"
		}
	}
	return "0/1"
}

func (ft *Table) GetLabelOperation(rowId string) map[string]Operation {
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

func (ft *Table) GetPodOperation() map[string]Operation {
	return map[string]Operation{
		"click": {Key: "goto", Target: "orgRoot"},
	}
}

func GetTableOperation() map[string]interface{} {

	res := map[string]interface{}{}
	for key, op := range ops {
		res[key] = interface{}(op)
	}
	return res
}

type State struct {
	IsFirstFilter   bool                   `json:"is_first_filter,omitempty"`
	PageNo          int                    `json:"page_no,omitempty"`
	PageSize        int                    `json:"page_size,omitempty"`
	Total           int                    `json:"total,omitempty"`
	Query           map[string]interface{} `json:"query,omitempty"`
	SelectedRowKeys []string               `json:"selected_row_keys,omitempty"`
	Start           time.Time              `json:"start,omitempty"`
	End             time.Time              `json:"end,omitempty"`
	Name            string                 `json:"name,omitempty"`
	ClusterName     string                 `json:"cluster_name,omitempty"`
	Namespace       string                 `json:"namespace,omitempty"`
	SortColumn      SorterData             `json:"sorter,omitempty"`
}

type SorterData struct {
	Field string           `json:"field"`
	Order common.OrderEnum `json:"order"`
}
