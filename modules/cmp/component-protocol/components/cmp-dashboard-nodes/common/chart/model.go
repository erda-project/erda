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

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var (
	Allocated       = "Allocated"
	Free_Allocate   = "Free-Allocate"
	Cannot_Allocate = "Cannot-Allocate"

	Memory = "memory"
	CPU    = "cpu"
	Pods   = "pods"

	DefaultFormat = "{d}%\n"
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

func GetScaleValue(quantity1 *resource.Quantity, quantity2 *resource.Quantity, quantity3 *resource.Quantity) (string, string, string) {
	factor := 10
	for ; (quantity1.Value() != 0 && quantity1.Value() > int64(1<<factor)) && ((quantity1.Value() != 0) && quantity2.Value() > int64(1<<factor)) && (quantity3.Value() != 0 && quantity3.Value() > int64(1<<factor)); factor += 10 {
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
	cht.SDK = cputil.SDK(ctx)
	cht.Props.Option.Series[0].Data = (*gs)[ResourceType+"Chart"].([]DataItem)
	return common.Transfer(cht.Props, &c.Props)
}

type Props struct {
	Option Option `json:"option"`
	Title  string `json:"title"`
	Style  Style  `json:"style"`
}

type Option struct {
	Color  []string `json:"color"`
	Legend Legend   `json:"legend"`
	Grid   Grid     `json:"grid"`
	Series []Serie  `json:"series"`
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
	Data   []string `json:"data"`
	Bottom string   `json:"bottom"`
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

func (cht *Chart) GetProps(name string) Props {
	return Props{
		Title: name,
		Option: Option{
			Color:  []string{"orange", "green", "red"},
			Legend: Legend{Data: []string{cht.SDK.I18n(Allocated), cht.SDK.I18n(Free_Allocate), cht.SDK.I18n(Cannot_Allocate)}, Bottom: "0"},
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
		},
		Style: Style{Flex: 1},
	}
}
