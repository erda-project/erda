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

package metrics

type MetricsRequest struct {
	UserID       string
	OrgID        string
	ClusterName  string
	ResourceType string
	ResourceKind string
	PodRequests  []MetricsPodRequest
	NodeRequests []MetricsNodeRequest
}

type MetricsPodRequest struct {
	PodName   string
	Namespace string
}

type MetricsNodeRequest struct {
	IP string
}

type MetricsData struct {
	// if qurey pod resource, used means usedPercent. request and total are useless.
	Used    float64 `json:"used"`
	Request float64 `json:"request"`
	Total   float64 `json:"total"`
}
