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
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/metrics"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	CpuUsageSelectStatement = `SELECT cpu_cores_usage ,n_cpus ,cpu_usage_active FROM status_page 
	WHERE cluster_name::tag=$cluster_name && hostname::tag=$hostname
	ORDER BY TIMESTAMP DESC`
	MemoryUsageSelectStatement = `SELECT mem_used , mem_total , mem_used_percent FROM status_page 
	WHERE cluster_name::tag=$cluster_name && hostname::tag=$hostname
	ORDER BY TIMESTAMP DESC`
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
}

func (t *Table) Init(ctx servicehub.Context) error { return nil }

func (t *Table) Run(ctx context.Context) error {
	return nil
}

type TableInterface interface {
	SetData(resources apistructs.SteveResource, resName v1.ResourceName) error
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
	ID      string             `json:"id"`
	Status  common.SteveStatus `json:"status"`
	Node    Node               `json:"node"`
	Role    string             `json:"role"`
	Version string             `json:"version"`
	//
	Distribution     Distribution         `json:"distribution"`
	Usage            Distribution         `json:"use"`
	DistributionRate Distribution         `json:"distribution_rate"`
	Labels           Labels               `json:"Labels"`
	Operations       map[string]Operation `json:"operations"`
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

type Node struct {
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
	RenderType string            `json:"render_type"`
	Value      DistributionValue `json:"value"`
	Status     common.UsageStatusEnum
}

type DistributionValue struct {
	Text    string `json:"text"`
	Percent int    `json:"percent"`
}

type LabelsValue struct {
	Label string `json:"label"`
	Group string `json:"group"`
}

func (t *Table) GetUsageValue(node *v1.Node, resName v1.ResourceName) (*DistributionValue, error) {
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
		Params: map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue(node.ClusterName),
			"hostname":     structpb.NewStringValue(node.Name),
		},
	}
	switch resName {
	case v1.ResourceCPU:
		req.Statement = CpuUsageSelectStatement
	case v1.ResourceMemory:
		req.Statement = MemoryUsageSelectStatement
	default:
		return nil, common.ResourceNotFoundErr
	}
	ctx, cancel := context.WithDeadline(t.Ctx, time.Now().Add(3*time.Second))
	defer cancel()
	select {
	case <-ctx.Done():
		logrus.Warningf("metrics service is busy")
		break
	default:
		if resp, err = t.Metric.Query(req, string(resName)); err != nil {
			return nil, err
		}
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

func (t *Table) GetItemStatus(node *v1.Node) (*common.SteveStatus, error) {
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

func (t *Table) GetDistributionValue(node *v1.Node, resName v1.ResourceName) (*DistributionValue, error) {
	var (
		pods       *apistructs.SteveCollection
		err        error
		containers []v1.Container
		resValue   resource.Quantity
		allocValue int64
	)
	req := &apistructs.SteveRequest{
		ClusterName:   node.ClusterName,
		Namespace:     node.Namespace,
		Type:          apistructs.K8SPod,
		LabelSelector: []string{fmt.Sprintf("=%s", node.Name)},
	}
	if pods, err = t.CtxBdl.Bdl.ListSteveResource(req); err != nil {
		return nil, err
	}
	for _, stevePod := range pods.Data {
		pod := v1.Pod{}
		if err = common.Transfer(stevePod, &pod); err != nil {
			return nil, err
		}
		for _, container := range pod.Spec.Containers {
			containers = append(containers, container)
		}
	}
	switch resName {
	case v1.ResourceCPU:
		for _, container := range containers {
			resValue.Add(container.Resources.Requests.Cpu().DeepCopy())
		}
		allocValue = node.Status.Allocatable.Cpu().Value()
	case v1.ResourceMemory:
		for _, container := range containers {
			resValue.Add(container.Resources.Requests.Memory().DeepCopy())
		}
		allocValue = node.Status.Allocatable.Memory().Value()
	default:
		return nil, common.ResourceNotFoundErr
	}
	allocDecimal := float64(allocValue)
	usageDecimal := float64(resValue.Value())
	return &DistributionValue{
		Text:    fmt.Sprintf("%.1f/%.1f", usageDecimal, allocDecimal),
		Percent: int(usageDecimal * 100 / allocDecimal),
	}, nil
}

func (t *Table) GetDistributionRate(node *v1.Node, resName v1.ResourceName) (*DistributionValue, error) {
	switch resName {
	case v1.ResourceCPU:
		cpuAllocatable := node.Status.Allocatable.Cpu().Value()
		cpuCapacity := node.Status.Capacity.Cpu().Value()
		return &DistributionValue{
			Text:    fmt.Sprintf("%d/%d", cpuAllocatable, cpuCapacity),
			Percent: common.GetPercent(float64(cpuAllocatable), float64(cpuCapacity)),
		}, nil
	case v1.ResourceMemory:
		cpuAllocatable := node.Status.Allocatable.Memory().Value()
		cpuCapacity := node.Status.Capacity.Memory().Value()
		return &DistributionValue{
			Text:    fmt.Sprintf("%d/%d", cpuAllocatable, cpuCapacity),
			Percent: common.GetPercent(float64(cpuAllocatable), float64(cpuCapacity)),
		}, nil
	default:
		return nil, common.ResourceNotFoundErr
	}
}

// SetComponentValue mapping CpuInfoTable properties to Component
func (t *Table) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(t.State); err != nil {
		return err
	}
	c.State = state
	return nil
}

// GetOpsInfo return request meta
func (t *Table) GetOpsInfo(opsData interface{}) (*Meta, error) {
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

func (t *Table) GetNodeAddress(addrs []v1.NodeAddress) string {
	for _, addr := range addrs {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}

func (t *Table) RenderChangePageSize(ops interface{}) error {
	meta, err := t.GetOpsInfo(ops)
	if err != nil {
		return err
	}
	t.State.PageNo = 1
	t.State.PageSize = meta.PageSize
	return nil
}

func (t *Table) RenderChangePageNo(ops interface{}) error {
	meta, err := t.GetOpsInfo(ops)
	if err != nil {
		return err
	}
	t.State.PageNo = meta.PageNo
	return nil
}
func (t *Table) GetRole(labels map[string]string) string {
	res := make([]string, 0)
	for k := range labels {
		if strings.HasPrefix(k, "node-role") {
			splits := strings.Split(k, "\\")
			res = append(res, splits[len(splits)-1])
		}
	}
	return strutil.Join(res, ",", true)
}

func (t *Table) GetPodLabels(labels map[string]string) []LabelsValue {
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

func (t *Table) GetLabelOperation(rowId string) map[string]Operation {
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

func (t *Table) GetNodeOperation() map[string]Operation {
	return map[string]Operation{
		"click": {Key: "goto", Target: "orgRoot"},
	}
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
	SortColumnName  string                 `json:"sorter,omitempty"`
	Asc             bool                   `json:"asc,omitempty"`
}
