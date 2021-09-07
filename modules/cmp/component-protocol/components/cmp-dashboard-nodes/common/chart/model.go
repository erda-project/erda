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

package chart

import (
	"context"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	"github.com/rancher/wrangler/pkg/data"
	"github.com/spf13/cast"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

var (
	Distributed_Desc = "已分配"
	Free_Desc        = "剩余分配"
	Locked_Desc      = "不可分配"

	Memory = "Memory"
	CPU    = "CPU"
	Pods   = "Pods"

	DefaultDegree = 60.0
	DefaultFormat = "{d}%\n{c}/60"
)

type Chart struct {
	base.DefaultProvider
	SDK    *cptype.SDK
	Ctx    context.Context
	CtxBdl *bundle.Bundle
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}
type ChartInterface interface {
	ChartRender(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error
}

func setData(nodes []data.Object, resourceName string) []DataItem {
	var allocatableTotal, capacityTotal, unAllocatableTotal float64
	if len(nodes) == 0{
		return []DataItem{}
	}
	for _, node := range nodes {
		allocatableTotal += cast.ToFloat64(node.String("extra", "parsedResource", "allocated", resourceName))
		capacityTotal += cast.ToFloat64(node.String("extra", "parsedResource", "capacity", resourceName))
		unAllocatableTotal += cast.ToFloat64(node.String("extra", "parsedResource", "unallocatable", resourceName))
	}
	return []DataItem{{
		Value: allocatableTotal / capacityTotal * DefaultDegree,
		Name:  Distributed_Desc,
		Label: Label{DefaultFormat},
	}, {
		Value: (capacityTotal - unAllocatableTotal - allocatableTotal) / capacityTotal * DefaultDegree,
		Name:  Free_Desc,
		Label: Label{DefaultFormat},
	}, {
		Value: unAllocatableTotal / capacityTotal * DefaultDegree,
		Name:  Locked_Desc,
		Label: Label{DefaultFormat},
	}}
}

func (cht *Chart) ChartRender(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData, ResourceType string) error {
	cht.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	var (
		nodes []data.Object
	)
	nodes = (*gs)["nodes"].([]data.Object)
	cht.Props.Option.Series[0].Data = setData(nodes, ResourceType)
	c.Props = cht.Props
	return nil
}

type Props struct {
	Option Option `json:"option"`
}

type Option struct {
	Color  []string `json:"color"`
	Legend Legend   `json:"legend"`
	Grid   Grid     `json:"grid"`
	Series []Serie  `json:"series"`
}

type Serie struct {
	Type   string     `json:"type"`
	Radius string     `json:"radius"`
	Data   []DataItem `json:"data"`
}

type Legend struct {
	Data []string `json:"data"`
}

type Grid struct {
	Bottom       int  `json:"bottom"`
	Top          int  `json:"top"`
	ContainLabel bool `json:"containLabel"`
}

type Style struct {
	Flex int `json:"flex"`
}

type DataItem struct {
	Value float64 `json:"value"`
	Name  string  `json:"name"`
	Label Label   `json:"label"`
}

type Label struct {
	Formatter string `json:"formatter"`
}

func (c *Chart) GetProps() Props {
	return Props{Option: Option{
		Color:  []string{"green", "orange", "red"},
		Legend: Legend{Data: []string{c.SDK.I18n("allocated"), c.SDK.I18n("cannot-allocated"), c.SDK.I18n("free-allocate")}},
		Grid: Grid{
			Bottom:       0,
			Top:          0,
			ContainLabel: true,
		},
		Series: []Serie{{
			Type:   "pie",
			Radius: "60%",
			Data:   nil,
		}},
	}}
}
