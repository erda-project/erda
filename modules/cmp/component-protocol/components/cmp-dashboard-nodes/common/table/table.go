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
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/label"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-pods/common/table"
)

type Table struct {
	TableInterface
	base.DefaultProvider
	CtxBdl         *bundle.Bundle
	SDK            *cptype.SDK
	Ctx            context.Context
	Type           string                    `json:"type"`
	Props          map[string]interface{}    `json:"props"`
	Operations     map[string]interface{}    `json:"operations"`
	BatchOperation map[string]BatchOperation `json:"batchOperation"`
	State          State                     `json:"state"`
}

type TableInterface interface {
	SetData(object data.Object, resName v1.ResourceName) error
}

type Columns struct {
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
	Width     int    `json:"width,omitempty"`
	Sortable  bool   `json:"sorter,omitempty"`
	Fixed     string `json:"fixed,omitempty"`
}

type Meta struct {
	Id       int    `json:"id,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	PageNo   int    `json:"pageNo,omitempty"`
	Sorter   Sorter `json:"sorter"`
}

type Sorter struct {
	Field string `json:"field,omitempty"`
	Order string `json:"order,omitempty"`
}

type Scroll struct {
	X int `json:"x,omitempty"`
}

type RowItem struct {
	ID      string             `json:"id"`
	Status  common.SteveStatus `json:"status"`
	Node    Node               `json:"node"`
	Role    string             `json:"role"`
	Version string             `json:"version"`
	//
	Distribution Distribution `json:"distribution"`
	Usage        Distribution `json:"usage"`
	UsageRate    Distribution `json:"usageRate"`
	Operate      Operate      `json:"operate"`
	BatchOptions []string     `json:"batchOptions"`
}

type Operate struct {
	RenderType string               `json:"renderType"`
	Operations map[string]Operation `json:"operations"`
}

type Operation struct {
	Key           string      `json:"key"`
	Reload        bool        `json:"reload"`
	FillMeta      string      `json:"fillMeta,omitempty"`
	Target        string      `json:"target,omitempty"`
	Meta          interface{} `json:"meta,omitempty"`
	ClickableKeys interface{} `json:"clickableKeys,omitempty"`
	Text          string      `json:"text"`
	Command       Command     `json:"command,omitempty"`
}
type BatchOperation struct {
	Key       string   `json:"key"`
	Text      string   `json:"text"`
	Reload    bool     `json:"reload"`
	ShowIndex []string `json:"showIndex"`
}
type Command struct {
	Key     string       `json:"key"`
	Command CommandState `json:"state"`
	Target  string       `json:"target"`
	JumpOut bool         `json:"jumpOut"`
}

type CommandState struct {
	Query    QueryData `json:"query"`
	Visible  bool      `json:"visible"`
	FormData FormData  `json:"formData"`
}
type QueryData struct {
	NodeId string `json:"nodeId"`
}

type FormData struct {
	RecordId string `json:"recordId"`
}

type Node struct {
	RenderType string        `json:"renderType"`
	Renders    []interface{} `json:"renders"`
}

type NodeLink struct {
	RenderType string               `json:"renderType"`
	Value      string               `json:"value"`
	Operation  map[string]Operation `json:"operation"`
	Reload     bool                 `json:"reload"`
}

type NodeTags struct {
	RenderType string               `json:"renderType"`
	Value      []label.Label        `json:"value"`
	Operation  map[string]Operation `json:"operation"`
}

type Labels struct {
	RenderType string               `json:"renderType"`
	Value      []label.Label        `json:"value"`
	Operation  map[string]Operation `json:"operation"`
}

type Distribution struct {
	RenderType string            `json:"renderType"`
	Value      DistributionValue `json:"value"`
	Status     string
}

type DistributionValue struct {
	Text    string `json:"text"`
	Percent string `json:"percent"`
}

type LabelsValue struct {
	Label string `json:"label"`
	Group string `json:"group"`
}

func (t *Table) GetUsageValue(metricsData apistructs.MetricsData) *DistributionValue {
	return &DistributionValue{
		Text:    fmt.Sprintf("%.1f/%.1f", metricsData.Used, metricsData.Total),
		Percent: common.GetPercent(metricsData.Used, metricsData.Total),
	}
}

func (t *Table) GetItemStatus(node data.Object) (*common.SteveStatus, error) {
	if node == nil {
		return nil, common.NodeNotFoundErr
	}
	ss := &common.SteveStatus{
		RenderType: "textWithBadge",
	}
	ss.Value = t.SDK.I18n(node.StringSlice("metadata", "fields")[1])
	ss.Status = node.StringSlice("metadata", "fields")[1]
	return ss, nil
}

func (t *Table) GetDistributionValue(metricsData apistructs.MetricsData) *DistributionValue {
	return &DistributionValue{
		Text:    fmt.Sprintf("%.1f/%.1f", metricsData.Request, metricsData.Total),
		Percent: common.GetPercent(metricsData.Request, metricsData.Total),
	}
}

func (t *Table) GetDistributionRate(metricsData apistructs.MetricsData) *DistributionValue {
	return &DistributionValue{
		Text:    fmt.Sprintf("%d/%d", metricsData.Used, metricsData.Request),
		Percent: common.GetPercent(metricsData.Request, metricsData.Total),
	}
}

// SetComponentValue mapping CpuInfoTable properties to Component
func (t *Table) SetComponentValue(c *cptype.Component) error {
	var (
		err error
	)
	if err = common.Transfer(t.State, &c.State); err != nil {
		return err
	}
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

func (t *Table) DeleteNode(nodeNames []string) error {
	for _, name := range nodeNames {
		req := &apistructs.SteveRequest{
			UserID:      t.SDK.Identity.UserID,
			OrgID:       t.SDK.Identity.OrgID,
			Type:        apistructs.K8SNode,
			ClusterName: t.SDK.InParams["clusterName"].(string),
			Name:        name,
		}
		err := t.CtxBdl.DeleteSteveResource(req)
		if err != nil {
			return err
		}
	}
	return nil
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

func (t *Table) GetTableOperation() map[string]interface{} {
	ops := map[string]table.Operation{
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

func (t *Table) GetTableBatchOperation() map[string]BatchOperation {
	ops := map[string]BatchOperation{
		"delete": {
			Key:    "delete",
			Reload: true,
			Text:   t.SDK.I18n("delete"),
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
	return ops
}

func (t *Table) GetPodLabels(labels data.Object) []label.Label {
	labelValues := make([]label.Label, 0)
	for key, value := range labels {
		lv := label.Label{
			Value: fmt.Sprintf("%s=%s", key, value),
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
					FormData: FormData{RecordId: rowId},
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
				"abc":      "",
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

func (t *Table) GetRenders(ip string, id string, labelMap data.Object) []interface{} {
	nl := NodeLink{
		RenderType: "linkText",
		Value:      ip,
		Operation: map[string]Operation{"click": {
			Key: "gotoNodeDetail",
			Command: Command{
				Key:    "goto",
				Target: "orgRoot",
			},
			Text:   t.SDK.I18n("nodeDetail"),
			Reload: false,
		},
		},
		Reload: false,
	}
	nt := Labels{
		RenderType: "tagsRow",
		Value:      t.GetPodLabels(labelMap),
		Operation:  t.GetLabelOperation(id),
	}
	return []interface{}{nl, nt}
}

func (t *Table) GetOperate(id string) Operate {
	return Operate{
		RenderType: "tableOperation",
		Operations: map[string]Operation{
			"gotoPod": {Key: "gotoPod", Command: Command{
				Key: "goto",
				Command: CommandState{
					Visible: false,
					Query:   QueryData{NodeId: id},
				},
				Target: "cmpClustersPods",
			}},
		},
	}
}

type State struct {
	PageNo          int      `json:"pageNo,omitempty"`
	PageSize        int      `json:"pageSize,omitempty"`
	Total           int      `json:"total,omitempty"`
	SelectedRowKeys []string `json:"selectedRowKeys,omitempty"`
	Sorter          Sorter   `json:"sorter"`
}
