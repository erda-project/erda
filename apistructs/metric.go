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

package apistructs

// HostMetricResponse 主机监控资源响应
type HostMetricResponse struct {
	Header
	Data HostMetricResponseData `json:"data"`
}

// HostMetricResponseData 主机监控资源响应数据结构
/*
"data": {
	"results": [{
		"data": [{
			"avg.load5": {
				"agg": "avg",
				"axisIndex": 0,
				"chartType": "",
				"data": 0.45545454545454545,
				"name": "load5",
				"tag": "10.168.0.101",
				"unit": "",
				"unitType": ""
			}
         }],
		"name": "system"
	}]
}
*/
type HostMetricResponseData struct {
	Results []HostMetricResult `json:"results"`
}

// HostMetricResult result结构
type HostMetricResult struct {
	Name string                  `json:"name"`
	Data []map[string]MetricData `json:"data"`
}

// MetricData metric结构
type MetricData struct {
	Tag  string  `json:"tag"` // host ip
	Name string  `json:"name"`
	Data float64 `json:"data"`
	Agg  string  `json:"agg"`
}

// HostMetric host metric bundle数据结构
type HostMetric struct {
	CPU    float64 // 百分比值， eg: 19%, 则cpu为19
	Memory float64 // 百分比
	Disk   float64 // 百分比
	Load   float64
}
