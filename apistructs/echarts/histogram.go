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

package echarts

type Histogram struct {
	XAxis  XAxis            `json:"xAxis"`
	YAxis  YAxis            `json:"yAxis"`
	Series []HistogramSerie `json:"series"`
	Name   string           `json:"name"`
}

type XAxis struct {
	Type string   `json:"type"`
	Data []string `json:"data"`
}

type YAxis struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type HistogramSerie struct {
	Data []float64 `json:"data"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

type Series struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	Data []float64 `json:"data"`
}
