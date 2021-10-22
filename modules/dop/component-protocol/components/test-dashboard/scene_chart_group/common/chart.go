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

package common

type BarProps struct {
	ChartType string `json:"chartType"`
	Option    Option `json:"option"`
	Title     string `json:"title"`
}

type Option struct {
	DataZoom []DataZoom `json:"dataZoom"`
	Grid     Grid       `json:"grid"`
	Series   []Series   `json:"series"`
	XAxis    []XAxis    `json:"xAxis"`
	YAxis    []YAxis    `json:"yAxis"`
}

type DataZoom struct {
	EndValue       int64  `json:"endValue,omitempty"`
	HandleSize     int64  `json:"handleSize,omitempty"`
	Orient         string `json:"orient,omitempty"`
	ShowDataShadow bool   `json:"showDataShadow,omitempty"`
	ShowDetail     bool   `json:"showDetail,omitempty"`
	StartValue     int64  `json:"startValue,omitempty"`
	Throttle       int64  `json:"throttle,omitempty"`
	Type           string `json:"type,omitempty"`
	Width          int64  `json:"width,omitempty"`
	ZoomLock       bool   `json:"zoomLock,omitempty"`
}

type Grid struct {
	Right int64 `json:"right"`
}

type Series struct {
	Data  []Data `json:"data"`
	Label Label  `json:"label"`
}

type Data struct {
	Operations map[string]interface{} `json:"operations"`
	Value      int64                  `json:"value"`
}

type Label struct {
	Show bool `json:"show"`
}

type Tooltip struct {
	Show    bool   `json:"show"`
	Trigger string `json:"trigger"`
}

type XAxis struct {
	Type string `json:"type"`
}

type YAxis struct {
	Type string   `json:"type"`
	Data []string `json:"data"`
}

func NewBarProps(values []int64, categories []string, title string) BarProps {
	data := make([]Data, 0, len(values))
	for _, v := range values {
		data = append(data, Data{
			Operations: nil,
			Value:      v,
		})
	}
	return BarProps{
		ChartType: "bar",
		Option: Option{
			DataZoom: []DataZoom{
				{
					EndValue:   int64(len(values)),
					Orient:     "vertical",
					StartValue: 10,
					Throttle:   0,
					Type:       "inside",
					ZoomLock:   true,
				},
				{
					EndValue:       int64(len(values)),
					HandleSize:     15,
					Orient:         "vertical",
					ShowDataShadow: false,
					ShowDetail:     false,
					StartValue:     10,
					Throttle:       0,
					Type:           "slider",
					Width:          15,
					ZoomLock:       true,
				},
			},
			Grid: Grid{
				Right: 30,
			},
			Series: []Series{
				{
					Data:  data,
					Label: Label{Show: true},
				},
			},
			XAxis: []XAxis{
				{Type: "value"},
			},
			YAxis: []YAxis{
				{
					Type: "category",
					Data: categories,
				},
			},
		},
		Title: title,
	}
}
