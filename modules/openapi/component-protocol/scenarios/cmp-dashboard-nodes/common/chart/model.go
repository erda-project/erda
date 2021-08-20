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

package chart

var (
	Distributed_Desc= "已分配"
	Free_Desc= "剩余分配"
	Locked_Desc= "不可分配"
)


type Chart struct {
	Type   string    `json:"type"`
	Data   ChartData `json:"data"`
	Props  Props     `json:"props"`
}

type ChartData struct {
	Results Result `json:"results"`
}

type Result struct {
	Data []ChartDataItem `json:"data"`
}

type Props struct {
	ChartType  string   `json:"chart_type"`
	LegendData []string `json:"legend_data"`
	Style      Style    `json:"style"`
}

type Style struct {
	Flex int `json:"flex"`
}

type ChartDataItem struct {
	Value float64 `json:"value"`
	Name  string   `json:"name"`
}

