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

package apistructs

type GaugeRequest struct {
	MemPerNode  uint64   `json:"memPerNode"`
	CpuPerNode  uint64   `json:"cpuPerNode"`
	ClusterName []string `json:"clusterName"`
}

type TableRequest struct {
	MemoryUnit  int
	CpuUnit     int
	ClusterName []string
}

type ClassRequest struct {
	ResourceType string
	ClusterName  []string
}

type TrendRequest struct {
	ResourceType string
	Start        int64
	End          int64
	Interval     string
	ProjectId    []string
	ClusterName  []string
}

type ResourceResp struct {
	MemRequest           float64
	CpuRequest           float64
	MemTotal             float64
	CpuTotal             float64
	CpuQuota             float64
	MemQuota             float64
	IrrelevantCpuRequest float64
	IrrelevantMemRequest float64
}
