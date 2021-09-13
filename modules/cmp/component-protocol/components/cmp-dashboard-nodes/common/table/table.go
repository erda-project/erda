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
	"encoding/base64"
	"encoding/json"
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
	Y int `json:"y,omitempty"`
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
	Params   Params                 `json:"params,omitempty"`
	Visible  bool                   `json:"visible,omitempty"`
	FormData FormData               `json:"formData,omitempty"`
	Query    map[string]interface{} `json:"query,omitempty"`
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

func (t *Table) GetUsageValue(metricsData apistructs.MetricsData, resourceType TableType) DistributionValue {
	return DistributionValue{
		Text:    t.GetScaleValue(metricsData.Used, metricsData.Total, resourceType),
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

func (t *Table) GetDistributionValue(metricsData apistructs.MetricsData, resourceType TableType) DistributionValue {
	return DistributionValue{
		Text:    t.GetScaleValue(metricsData.Request, metricsData.Total, resourceType),
		Percent: common.GetPercent(metricsData.Request, metricsData.Total),
	}
}

func (t *Table) GetDistributionRate(metricsData apistructs.MetricsData, resourceType TableType) DistributionValue {
	return DistributionValue{
		Text:    t.GetScaleValue(metricsData.Used, metricsData.Request, resourceType),
		Percent: common.GetPercent(metricsData.Used, metricsData.Request),
	}
}

func (t *Table) GetScaleValue(a, b float64, resourceType TableType) string {
	level := []string{"", "K", "M", "G", "T"}
	i := 0
	switch resourceType {
	case Memory:
		for ; a > 1024 && b > 1024 && i < 4; i++ {
			a /= 1024
			b /= 1024
		}
		return fmt.Sprintf("%.1f%si/%.1f%si", a, level[i], b, level[i])
	case Cpu:
		for a > 1000 && b > 1000 && i < 4 {
			a /= 1000
			b /= 1000
		}
		return fmt.Sprintf("%.3f/%.3f", a, b)
	default:
		for ; a > 1000 && b > 1000 && i < 4; i++ {
			a /= 1000
			b /= 1000
		}
		return fmt.Sprintf("%d%s/%d%s", int64(a), level[i], int64(b), level[i])
	}
}

// SetComponentValue mapping properties to Component
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

func (t *Table) CordonNode(nodeNames []string) error {
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

func (t *Table) UncordonNode(nodeNames []string) error {
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
		"cordon": {
			Key:    "cordon",
			Reload: true,
			Text:   t.SDK.I18n("cordon"),
		},
		"uncordon": {
			Key:    "uncordon",
			Text:   t.SDK.I18n("uncordon"),
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
		l := fmt.Sprintf("%s=%s", key, value)
		group, displayName := t.GetLabelGroupAndDisplayName(l)
		lv := label.Label{
			Value: l,
			Name:  displayName,
			Group: group,
		}
		labelValues = append(labelValues, lv)
	}
	sort.Slice(labelValues, func(i, j int) bool {
		return labelValues[i].Group < labelValues[j].Group
	})
	return labelValues
}

func (t *Table) GetLabelGroupAndDisplayName(label string) (string, string) {
	//ls := []string{
	//	"dice/workspace-dev", "dice/workspace-test", "staging", "prod", "stateful", "stateless", "packJob", "cluster-service", "mono", "cordon", "drain", "platform",
	//}
	groups := make(map[string]string)
	groups["dice/workspace-dev=true"] = t.SDK.I18n("env")
	groups["dice/workspace-test=true"] = t.SDK.I18n("env")
	groups["dice/workspace-staging=true"] = t.SDK.I18n("env")
	groups["dice/workspace-prod=true"] = t.SDK.I18n("env")

	groups["dice/stateful-service=true"] = t.SDK.I18n("service")
	groups["dice/stateless-service=true"] = t.SDK.I18n("service")
	groups["dice/location-cluster-service=true"] = t.SDK.I18n("service")

	groups["dice/job=true"] = t.SDK.I18n("job")
	groups["dice/bigdata-job=true"] = t.SDK.I18n("job")

	groups["dice/lb=true"] = t.SDK.I18n("other")
	groups["dice/platform=true"] = t.SDK.I18n("other")

	if group, ok := groups[label]; ok {
		idx := strings.Index(label, "=true")
		return t.SDK.I18n(group), t.SDK.I18n(label[5:idx])
	}

	if strings.HasPrefix(label, "dice/org-") && strings.HasSuffix(label, "=true") {
		idx := strings.Index(label, "=true")
		return t.SDK.I18n("organization"), t.SDK.I18n(label[5:idx])
	}
	otherDisplayName := label
	if label == "dice/lb=true" || label == "dice/platform=true" {
		idx := strings.Index(label, "=true")
		otherDisplayName = t.SDK.I18n(label[5:idx])
	}
	return t.SDK.I18n("other"), otherDisplayName
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
			Reload:   true,
			FillMeta: "dlabel",
			Meta: map[string]interface{}{
				"recordId": rowId,
				"dlabel":   label.Label{Value: ""},
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
					Query:  map[string]interface{}{"nodeIP": ip},
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
}

func (t *Table) GetOperate(id string) Operate {
	obj := map[string]interface{}{
		"node": []string{
			id,
		},
	}
	data, _ := json.Marshal(obj)
	encode := base64.StdEncoding.EncodeToString(data)
	return Operate{
		RenderType: "tableOperation",
		Operations: map[string]Operation{
			"gotoPod": {Key: "gotoPod", Command: Command{
				Key: "goto",
				Command: CommandState{
					Query: map[string]interface{}{
						"filter__urlQuery": encode,
					},
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
