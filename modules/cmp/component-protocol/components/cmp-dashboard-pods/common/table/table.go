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
	"strings"

	"github.com/rancher/wrangler/pkg/data"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	common "github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-pods/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

type Table struct {
	base.DefaultProvider
	TableInterface
	Ctx        context.Context
	CtxBdl     *bundle.Bundle
	SDK        *cptype.SDK
	Type       string                 `json:"type"`
	Props      map[string]interface{} `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
}

type TableInterface interface {
	SetData(object data.Object, resName v1.ResourceName) error
}
type CpuTable struct {
	Data       map[string][]Data    `json:"data"`
	Props      Props                `json:"props"`
	Operations map[string]Operation `json:"operations"`
	Type       string               `json:"type"`
	State      State                `json:"state"`
}

type Props struct {
	PageSizeOptions string  `json:"pageSizeOptions"`
	Columns         Columns `json:"columns"`
	RowKey          string  `json:"rowKey"`
}

type State struct {
	PageNo     int        `json:"pageNo"`
	PageSize   int        `json:"pageSize"`
	Total      int        `json:"total"`
	SorterData SorterData `json:"sorterData"`
}

type SorterData struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type Data struct {
	Ready     string  `json:"ready"`
	Id        string  `json:"id"`
	Status    Status  `json:"status"`
	Namespace string  `json:"namespace"`
	Used      string  `json:"cpuUsed"`
	Percent   Percent `json:"cpuPercent"`
	Name      Name    `json:"name"`
	Ip        string  `json:"ip"`
	CpuLimit  string  `json:"cpuLimit"`
}

type Columns struct {
	DataIndex string `json:"dataIndex"`
	Title     string `json:"title"`
	Width     int    `json:"width"`
	Sorter    bool   `json:"sorter"`
}

type ChangePageNo struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type ChangeSort struct {
	Reload bool   `json:"reload"`
	Key    string `json:"key"`
}

type Status struct {
	RenderType  string      `json:"renderType"`
	Value       string      `json:"value"`
	StyleConfig StyleConfig `json:"styleConfig"`
	Tip         string      `json:"tip,omitempty"`
}

type Percent struct {
	RenderType string `json:"renderType"`
	Value      string `json:"value"`
	Tip        string `json:"tip"`
	Status     string `json:"status"`
}

type Name struct {
	Operations map[string]Operation `json:"operations"`
	RenderType string               `json:"renderType"`
	Value      string               `json:"value"`
}

type StyleConfig struct {
	Color string `json:"color"`
}

type Operation struct {
	Key     string  `json:"key"`
	Command Command `json:"command"`
	Reload  bool    `json:"reload"`
}

type Command struct {
	Key     string       `json:"key"`
	State   CommandState `json:"state"`
	Target  string       `json:"target"`
	JumpOut bool         `json:"jumpOut"`
}

type CommandState struct {
	Params Params `json:"params"`
}

type Params struct {
	PodId string `json:"podId"`
}

type Distribution struct {
	RenderType string            `json:"renderType"`
	Value      DistributionValue `json:"value"`
	Status     common.UsageStatusEnum
}

type DistributionValue struct {
	Text    string  `json:"text"`
	Percent float64 `json:"percent"`
}

type Meta struct {
	Id         int    `json:"id,omitempty"`
	PageSize   int    `json:"pageSize,omitempty"`
	PageNo     int    `json:"pageNo,omitempty"`
	SortColumn string `json:"sorter"`
}

type RowItem struct {
	Name        Name         `json:"name"`
	ID          string       `json:"id"`
	Status      Status       `json:"status"`
	Namespace   string       `json:"namespace"`
	IP          string       `json:"ip"`
	Request     string       `json:"request"`
	UsedPercent Distribution `json:"usedPercent"`
	Limit       string       `json:"limit"`
	Ready       string       `json:"ready"`
}

func (t *Table) GetUsageValue(metricsData apistructs.MetricsData) *DistributionValue {
	return &DistributionValue{
		Text:    fmt.Sprintf("%.1f/%.1f", metricsData.Used, metricsData.Total),
		Percent: common.GetPercent(metricsData.Used, metricsData.Total),
	}
}

func (t *Table) GetItemStatus(percent float64) *Status {
	ss := &Status{
		RenderType:  "progress",
		Value:       fmt.Sprintf("%f", percent),
		StyleConfig: StyleConfig{Color: getColor(percent)},
	}
	return ss
}

func getColor(percent float64) string {
	if percent > 90 {
		return common.ColorMap["red"]
	} else if percent > 60 {
		return common.ColorMap["orange"]
	}
	return common.ColorMap["green"]
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
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(t.State); err != nil {
		return err
	}
	c.State = state
	return nil
}

func (t *Table) GetResourceReq(pod data.Object, resourceKind string, resourceType v1.ResourceName) string {
	var format resource.Format
	if resourceType == "cpu" {
		format = resource.DecimalSI
	} else {
		format = resource.BinarySI
	}
	result := resource.NewQuantity(0, format)

	for _, container := range pod.Slice("spec", "containers") {
		result.Add(*parseResource(container.String("resources", resourceKind, string(resourceType)), format))
	}
	return result.String()
}

func parseResource(str string, format resource.Format) *resource.Quantity {
	if str == "" {
		return resource.NewQuantity(0, format)
	}
	res, _ := resource.ParseQuantity(str)
	return &res
}

func (t *Table) GetRole(labels []string) string {
	res := make([]string, 0)
	for _, k := range labels {
		if strings.HasPrefix(k, "node-role") {
			splits := strings.Split(k, "\\")
			res = append(res, splits[len(splits)-1])
		}
	}
	return strutil.Join(res, ",", true)
}

func (t *Table) GetIp(node data.Object) string {
	for _, k := range node.Slice("status", "addresses") {
		if k.String("type") == "InternalIP" {
			return k.String("address")
		}
	}
	return ""
}
