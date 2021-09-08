// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package table

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/filter"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/label"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/nodeFilter"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type Table struct {
	TableComponent GetRowItem
	base.DefaultProvider
	CtxBdl     *bundle.Bundle
	SDK        *cptype.SDK
	Ctx        context.Context
	Type       string                 `json:"type"`
	Props      map[string]interface{} `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
}

type TableType string

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1

	Pod    TableType = "pod"
	Memory TableType = "memory"
	Cpu    TableType = "cpu"
)

type Columns struct {
	Title     string `json:"title,omitempty"`
	DataIndex string `json:"dataIndex,omitempty"`
	Width     int    `json:"width,omitempty"`
	Sortable  bool   `json:"sorter,omitempty"`
	Fixed     string `json:"fixed,omitempty"`
}

type Meta struct {
	Id       int    `json:"id,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	PageNo   int    `json:"pageNo,omitempty"`
	Sorter   Sorter `json:"sorter,omitempty"`
}

type Sorter struct {
	Field string `json:"field,omitempty"`
	Order string `json:"order,omitempty"`
}

type Scroll struct {
	X int `json:"x,omitempty"`
}

type RowItem struct {
	ID      string      `json:"id,omitempty"`
	IP      string      `json:"IP,omitempty"`
	Status  SteveStatus `json:"Status,omitempty"`
	Node    Node        `json:"Node,omitempty"`
	Role    string      `json:"Role,omitempty"`
	Version string      `json:"Version,omitempty"`
	//
	Distribution Distribution `json:"Distribution,omitempty"`
	Usage        Distribution `json:"Usage,omitempty"`
	UsageRate    Distribution `json:"UsageRate,omitempty"`
	Operate      Operate      `json:"Operate,omitempty"`
	// batchOperations for json
	BatchOperations []string `json:"batchOperations,omitempty"`
}

type Operate struct {
	RenderType string               `json:"renderType,omitempty"`
	Operations map[string]Operation `json:"operations,omitempty"`
}

type Operation struct {
	Key           string      `json:"key,omitempty"`
	Reload        bool        `json:"reload"`
	FillMeta      string      `json:"fillMeta,omitempty"`
	Target        string      `json:"target,omitempty"`
	Meta          interface{} `json:"meta,omitempty"`
	ClickableKeys interface{} `json:"clickableKeys,omitempty"`
	Text          string      `json:"text,omitempty"`
	Command       Command     `json:"command"`
}
type BatchOperation struct {
	Key       string   `json:"key,omitempty"`
	Text      string   `json:"text,omitempty"`
	Reload    bool     `json:"reload,omitempty"`
	ShowIndex []string `json:"showIndex,omitempty"`
}
type Command struct {
	Key     string       `json:"key,omitempty"`
	Command CommandState `json:"state,omitempty"`
	Target  string       `json:"target,omitempty"`
	JumpOut bool         `json:"jumpOut,omitempty"`
}

type CommandState struct {
	Params   Params   `json:"params,omitempty"`
	Visible  bool     `json:"visible,omitempty"`
	FormData FormData `json:"formData,omitempty"`
}
type Params struct {
	NodeId string `json:"nodeId,omitempty"`
	NodeIP string `json:"nodeIP,omitempty"`
}

type FormData struct {
	RecordId string `json:"recordId,omitempty"`
}

type Node struct {
	RenderType string        `json:"renderType,omitempty"`
	Renders    []interface{} `json:"renders,omitempty"`
}

type NodeLink struct {
	RenderType string               `json:"renderType,omitempty"`
	Value      string               `json:"value,omitempty"`
	Operations map[string]Operation `json:"operations,omitempty"`
	Reload     bool                 `json:"reload"`
}

type NodeTags struct {
	RenderType string               `json:"renderType,omitempty"`
	Value      []label.Label        `json:"value,omitempty"`
	Operation  map[string]Operation `json:"operation,omitempty"`
}

type Labels struct {
	RenderType string               `json:"renderType,omitempty"`
	Value      []label.Label        `json:"value,omitempty"`
	Operations map[string]Operation `json:"operations,omitempty"`
}

type Distribution struct {
	RenderType string `json:"renderType,omitempty"`
	//Value      DistributionValue `json:"value,omitempty"`
	Value  string `json:"value"`
	Status string `json:"status,omitempty"`
	Tip    string `json:"tip,omitempty"`
}

type DistributionValue struct {
	Text    string `json:"text"`
	Percent string `json:"percent"`
}

type LabelsValue struct {
	Label string `json:"label"`
	Group string `json:"group"`
}

type GetRowItem interface {
	GetRowItem(c data.Object, resName TableType) (*RowItem, error)
}

func (t *Table) GetUsageValue(metricsData apistructs.MetricsData) DistributionValue {
	return DistributionValue{
		Text:    fmt.Sprintf("%.1f/%.1f", metricsData.Used, metricsData.Total),
		Percent: common.GetPercent(metricsData.Used, metricsData.Total),
	}
}

func GetDistributionStatus(str string) string {
	d := cast.ToFloat64(str)
	if d >= 100 {
		return "error"
	} else if d >= 80 {
		return "warning"
	} else {
		return "success"
	}
}

func (t *Table) GetItemStatus(node data.Object) (*SteveStatus, error) {
	if node == nil {
		return nil, common.NodeNotFoundErr
	}
	ss := &SteveStatus{
		RenderType: "textWithBadge",
	}
	strs := make([]string, 0)
	for _, s := range strings.Split(node.StringSlice("metadata", "fields")[1], ",") {
		strs = append(strs, t.SDK.I18n(s))
	}
	ss.Value = strings.Join(strs, ",")
	if node.StringSlice("metadata", "fields")[1] == "Ready" {
		ss.Status = "success"
	} else {
		ss.Status = "error"
	}
	return ss, nil
}

func (t *Table) GetDistributionValue(metricsData apistructs.MetricsData) DistributionValue {
	return DistributionValue{
		Text:    fmt.Sprintf("%.1f/%.1f", metricsData.Request, metricsData.Total),
		Percent: common.GetPercent(metricsData.Request, metricsData.Total),
	}
}

func (t *Table) GetDistributionRate(metricsData apistructs.MetricsData) DistributionValue {
	return DistributionValue{
		Text:    fmt.Sprintf("%f/%f", metricsData.Used, metricsData.Request),
		Percent: common.GetPercent(metricsData.Used, metricsData.Request),
	}
}

// SetComponentValue mapping CpuInfoTable properties to Component
func (t *Table) SetComponentValue(c *cptype.Component) error {
	var err error
	if err = common.Transfer(t.State, &c.State); err != nil {
		return err
	}
	if err = common.Transfer(t.Props, &c.Props); err != nil {
		return err
	}
	if err = common.Transfer(t.Operations, &c.Operations); err != nil {
		return err
	}
	return nil
}

func (t *Table) RenderList(component *cptype.Component, tableType TableType, nodes []data.Object) error {
	var (
		err        error
		sortColumn string
		asc        bool
		items      []RowItem
	)
	if t.State.PageNo == 0 {
		t.State.PageNo = DefaultPageNo
	}
	if t.State.PageSize == 0 {
		t.State.PageSize = DefaultPageSize
	}
	if t.State.Sorter.Field != "" {
		sortColumn = t.State.Sorter.Field
		asc = strings.ToLower(t.State.Sorter.Order) == "ascend"
	}

	if items, err = t.SetData(nodes, tableType); err != nil {
		return err
	}
	if sortColumn != "" {
		refCol := reflect.ValueOf(RowItem{}).FieldByName(sortColumn)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			SortByString(items, sortColumn, asc)
		case reflect.TypeOf(Node{}):
			SortByNode(items, sortColumn, asc)
		case reflect.TypeOf(Distribution{}):
			SortByDistribution(items, sortColumn, asc)
		case reflect.TypeOf(SteveStatus{}):
			SortByStatus(items, sortColumn, asc)
		default:
			logrus.Errorf("sort by column %s not found", sortColumn)
			return common.TypeNotAvailableErr
		}
	}

	t.State.Total = len(nodes)
	start := (t.State.PageNo - 1) * t.State.PageSize
	end := mathutil.Min(t.State.PageNo*t.State.PageSize, t.State.Total)

	component.Data = map[string]interface{}{"list": items[start:end]}
	return nil
}

// SetData assemble rowItem of table
func (t *Table) SetData(nodes []data.Object, tableType TableType) ([]RowItem, error) {
	var (
		list []RowItem
		ri   *RowItem
		err  error
	)
	for i := 0; i < len(nodes); i++ {
		if ri, err = t.TableComponent.GetRowItem(nodes[i], tableType); err != nil {
			return nil, err
		}
		list = append(list, *ri)
	}
	return list, nil
}

func (t *Table) GetNodes(gs *cptype.GlobalStateData) ([]data.Object, error) {
	var nodes []data.Object
	if (*gs)["nodes"] == nil {
		// Get all nodes by cluster name
		nodeReq := &apistructs.SteveRequest{}
		nodeReq.OrgID = t.SDK.Identity.OrgID
		nodeReq.UserID = t.SDK.Identity.UserID
		nodeReq.Type = apistructs.K8SNode
		if t.SDK.InParams["clusterName"] != nil {
			nodeReq.ClusterName = t.SDK.InParams["clusterName"].(string)
		} else {
			return nil, common.ClusterNotFoundErr
		}
		resp, err := t.CtxBdl.ListSteveResource(nodeReq)
		if err != nil {
			return nil, err
		}
		nodes = resp.Slice("data")
		nodeFilter.DoFilter(nodes, t.State.Values)
	} else {
		nodes = (*gs)["nodes"].([]data.Object)
	}
	return nodes, nil
}

func (t *Table) FreezeNode(nodeNames []string) error {
	for _, name := range nodeNames {
		req := &apistructs.SteveRequest{
			UserID:      t.SDK.Identity.UserID,
			OrgID:       t.SDK.Identity.OrgID,
			Type:        apistructs.K8SNode,
			ClusterName: t.SDK.InParams["clusterName"].(string),
			Name:        name,
		}
		err := t.CtxBdl.CordonNode(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) UnFreezeNode(nodeNames []string) error {
	for _, name := range nodeNames {
		req := &apistructs.SteveRequest{
			UserID:      t.SDK.Identity.UserID,
			OrgID:       t.SDK.Identity.OrgID,
			Type:        apistructs.K8SNode,
			ClusterName: t.SDK.InParams["clusterName"].(string),
			Name:        name,
		}
		err := t.CtxBdl.UnCordonNode(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) GetNodeAddress(addrs []v1.NodeAddress) string {
	for _, addr := range addrs {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}

func (t *Table) GetTableOperation() map[string]interface{} {
	ops := map[string]Operation{
		"changePageNo": {
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": {
			Key:    "changePageSize",
			Reload: true,
		},
		"changeSort": {
			Key:    "changeSort",
			Reload: true,
		},
		"freeze": {
			Key:    "freeze",
			Reload: true,
			Text:   t.SDK.I18n("freeze"),
		},
		"unfreeze": {
			Key:    "unfreeze",
			Text:   t.SDK.I18n("unfreeze"),
			Reload: true,
		},
	}
	res := map[string]interface{}{}
	for key, op := range ops {
		res[key] = interface{}(op)
	}
	return res
}

func (t *Table) GetNodeLabels(labels data.Object) []label.Label {
	labelValues := make([]label.Label, 0)
	for key, value := range labels {
		lv := label.Label{
			Value: fmt.Sprintf("%s=%s", key, value),
			Group: t.GetLabelGroup(key),
		}
		labelValues = append(labelValues, lv)
	}
	return labelValues
}

func (t *Table) GetLabelGroup(label string) string {
	ls := []string{
		"dev", "test", "staging", "prod", "stateful", "stateless", "packJob", "cluster-service", "mono", "cordon", "drain", "platform",
	}
	groups := make(map[string]string)
	groups["dev"] = "env"
	groups["test"] = "env"
	groups["staging"] = "env"
	groups["prod"] = "env"

	groups["stateful"] = "service"
	groups["stateless"] = "service"

	groups["packJob"] = "packjob"

	groups["cluster-service"] = "other"
	groups["mono"] = "other"
	groups["cordon"] = "other"
	groups["drain"] = "other"
	groups["platform"] = "other"

	for _, l := range ls {
		if strings.Contains(label, l) {
			return groups[l]
		}
	}
	return "custom"
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
					FormData: FormData{RecordId: rowId},
				},
				Target: "addLabelModal",
			},
		},
		"delete": {
			Key:      "deleteLabel",
			Reload:   false,
			FillMeta: "label",
			Meta: map[string]string{
				"recordId": rowId,
				"label":    "",
			},
		},
	}
}
func (t *Table) GetIp(node data.Object) string {
	for _, k := range node.Slice("status", "addresses") {
		if k.String("type") == "InternalIP" {
			return k.String("address")
		}
	}
	return ""
}

func (t *Table) GetRenders(id, ip string, labelMap data.Object) []interface{} {
	nl := NodeLink{
		RenderType: "linkText",
		Value:      id,
		Operations: map[string]Operation{"click": {
			Key: "gotoNodeDetail",
			Command: Command{
				Key:    "goto",
				Target: "cmpClustersNodeDetail",
				Command: CommandState{
					Params: Params{NodeId: id, NodeIP: ip},
				},
				JumpOut: true,
			},
			Text:   t.SDK.I18n("nodeDetail"),
			Reload: false,
		},
		},
		Reload: false,
	}
	nt := Labels{
		RenderType: "tagsRow",
		Value:      t.GetNodeLabels(labelMap),
		Operations: t.GetLabelOperation(id),
	}
	return []interface{}{[]interface{}{nl}, []interface{}{nt}}

	//return []interface{}{nl, nt}
}

func (t *Table) GetOperate(id string) Operate {
	return Operate{
		RenderType: "tableOperation",
		Operations: map[string]Operation{
			"gotoPod": {Key: "gotoPod", Command: Command{
				Key: "goto",
				Command: CommandState{
					Params: Params{NodeId: id},
				},
				JumpOut: true,
				Target:  "cmpClustersPods",
			},
				Text:   t.SDK.I18n("查看") + "pods",
				Reload: false,
			},
		},
	}
}

// SortByString sort by string value
func SortByString(data []RowItem, sortColumn string, asc bool) {
	sort.Slice(data, func(i, j int) bool {
		a := reflect.ValueOf(data[i])
		b := reflect.ValueOf(data[j])
		if asc {
			return a.FieldByName(sortColumn).String() < b.FieldByName(sortColumn).String()
		}
		return a.FieldByName(sortColumn).String() > b.FieldByName(sortColumn).String()
	})
}

// SortByNode sort by node struct
func SortByNode(data []RowItem, _ string, asc bool) {
	sort.Slice(data, func(i, j int) bool {
		if asc {
			return data[i].Node.Renders[0].([]interface{})[0].(NodeLink).Value < data[j].Node.Renders[0].([]interface{})[0].(NodeLink).Value
		}
		return data[i].Node.Renders[0].([]interface{})[0].(NodeLink).Value > data[j].Node.Renders[0].([]interface{})[0].(NodeLink).Value
	})
}

// SortByDistribution sort by percent
func SortByDistribution(data []RowItem, sortColumn string, asc bool) {
	sort.Slice(data, func(i, j int) bool {
		a := reflect.ValueOf(data[i])
		b := reflect.ValueOf(data[j])
		aValue := cast.ToFloat64(a.FieldByName(sortColumn).FieldByName("Value").String())
		bValue := cast.ToFloat64(b.FieldByName(sortColumn).FieldByName("Value").String())
		if asc {
			return aValue < bValue
		}
		return aValue > bValue
	})
}

// SortByStatus sort by percent
func SortByStatus(data []RowItem, _ string, asc bool) {
	sort.Slice(data, func(i, j int) bool {
		if asc {
			return data[i].Status.Value < data[j].Status.Value
		}
		return data[i].Status.Value > data[j].Status.Value
	})
}

type State struct {
	PageNo          int           `json:"pageNo,omitempty"`
	PageSize        int           `json:"pageSize,omitempty"`
	Total           int           `json:"total,omitempty"`
	SelectedRowKeys []string      `json:"selectedRowKeys,omitempty"`
	Sorter          Sorter        `json:"sorterData,omitempty"`
	Values          filter.Values `json:"values"`
}

type SteveStatus struct {
	Value      string `json:"value,omitempty"`
	RenderType string `json:"renderType"`
	Status     string `json:"status"`
}
