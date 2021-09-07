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

	"github.com/rancher/wrangler/pkg/data"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var (
	Distributed_Desc = "已分配"
	Free_Desc        = "剩余分配"
	Locked_Desc      = "不可分配"

	Memory = "Memory"
	CPU    = "CPU"
	Pods   = "Pods"

	DefaultFormat = "{d}%\n{c}"
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
	//var allocatableTotal, capacityTotal, unAllocatableTotal float64
	resourceType := resource.DecimalSI
	if resourceName == Memory {
		resourceType = resource.BinarySI
	}
	allocatableQuantity := resource.NewQuantity(0, resourceType)
	capacityQuantity := resource.NewQuantity(0, resourceType)
	unAllocatableQuantity := resource.NewQuantity(0, resourceType)
	if len(nodes) == 0 {
		return []DataItem{}
	}
	for _, node := range nodes {
		allocatableQuantity.Add(*parseResource(node.String("extra", "parsedResource", "allocated", resourceName), resourceType))
		capacityQuantity.Add(*parseResource(node.String("extra", "parsedResource", "capacity", resourceName), resourceType))
		unAllocatableQuantity.Add(*parseResource(node.String("extra", "parsedResource", "unallocatable", resourceName), resourceType))
	}
	allocatableQuantity.ToUnstructured()
	capacityQuantity.Sub(*unAllocatableQuantity)
	capacityQuantity.Sub(*allocatableQuantity)
	allocatableStr, unAllocatableStr, capacityStr := GetScaleValue(allocatableQuantity, unAllocatableQuantity, capacityQuantity)
	return []DataItem{{
		Value: float64(allocatableQuantity.Value()),
		Name:  Distributed_Desc,
		Label: Label{Formatter: allocatableStr},
	}, {
		Value: float64(capacityQuantity.Value()),
		Name:  Free_Desc,
		Label: Label{capacityStr},
	}, {
		Value: float64(unAllocatableQuantity.Value()),
		Name:  Locked_Desc,
		Label: Label{unAllocatableStr},
	}}
}

func GetScaleValue(quantity1 *resource.Quantity, quantity2 *resource.Quantity, quantity3 *resource.Quantity) (string, string, string) {
	factor := 10
	for ; quantity1.Value() > int64(1<<factor) && quantity2.Value() > int64(1<<factor) && quantity3.Value() > int64(1<<factor); factor += 10 {
	}
	factor -= 10
	quantity1.Set(quantity1.Value() / (1 << factor))
	quantity2.Set(quantity2.Value() / (1 << factor))
	quantity3.Set(quantity3.Value() / (1 << factor))
	switch factor {
	case 0:
		return quantity1.String(), quantity2.String(), quantity3.String()
	case 10:
		return quantity1.String() + "K", quantity2.String() + "K", quantity3.String() + "K"
	case 20:
		return quantity1.String() + "M", quantity2.String() + "M", quantity3.String() + "M"
	case 30:
		return quantity1.String() + "G", quantity2.String() + "G", quantity3.String() + "G"
	case 40:
		return quantity1.String() + "T", quantity2.String() + "T", quantity3.String() + "T"
	}
	return "", "", ""
}

func parseResource(str string, format resource.Format) *resource.Quantity {
	if str == "" {
		return resource.NewQuantity(0, format)
	}
	res, _ := resource.ParseQuantity(str)
	return &res
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
	Style  Style  `json:"style"`
}

type Option struct {
	Color  []string `json:"color"`
	Legend Legend   `json:"legend"`
	Grid   Grid     `json:"grid"`
	Series []Serie  `json:"series"`
	Title  Title    `json:"title"`
}

type Title struct {
	Text      string    `json:"text"`
	TextStyle TextStyle `json:"textStyle"`
}

type TextStyle struct {
	FrontSize int `json:"frontSize"`
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

func (c *Chart) GetProps(name string) Props {
	return Props{Option: Option{
		Color:  []string{"#F7A76B", "#6CB38B", "#DE5757"},
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
		Title: Title{
			Text:      name,
			TextStyle: TextStyle{FrontSize: 14},
		},
	}, Style: Style{Flex: 1}}
}
