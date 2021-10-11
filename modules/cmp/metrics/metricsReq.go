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

import "github.com/erda-project/erda/modules/cmp/cache"

type MetricsRequest struct {
	UserId       string
	OrgId        string
	Cluster      string
	Type         string
	Kind         string
	PodRequests  []MetricsPodRequest
	NodeRequests []MetricsNodeRequest
}

func (m MetricsRequest) UserID() string {
	return m.UserId
}

func (m MetricsRequest) OrgID() string {
	return m.OrgId
}

func (m MetricsRequest) ResourceType() string {
	return m.Type
}

func (m MetricsRequest) ResourceKind() string {
	return m.Kind
}

func (m MetricsRequest) ClusterName() string {
	return m.Cluster
}

type MetricsPodRequest struct {
	*MetricsRequest
	Name         string
	PodNamespace string
}

func (m MetricsPodRequest) CacheKey() string {
	return cache.GenerateKey(m.ClusterName(), m.PodName(), m.Namespace(), m.ResourceType(), m.ResourceKind())
}

func (m *MetricsPodRequest) ResourceType() string {
	return m.Type
}

func (m *MetricsPodRequest) ResourceKind() string {
	return m.Kind
}

func (m *MetricsPodRequest) ClusterName() string {
	return m.Cluster
}

func (m *MetricsPodRequest) PodName() string {
	return m.Name
}

func (m *MetricsPodRequest) Namespace() string {
	return m.PodNamespace
}

type Basic interface {
	UserID() string
	OrgID() string
	ResourceType() string
	ResourceKind() string
	ClusterName() string
}
type NodeMetrics interface {
	Basic
	IP() string
}

type PodMetrics interface {
	Basic
	PodName() string
	Namespace() string
}

type Key interface {
	CacheKey() string
}

type MetricsReqInterface interface {
	Key
	PodMetrics
	NodeMetrics
}

type MetricsNodeRequest struct {
	*MetricsRequest
	Ip string
}

func (m *MetricsNodeRequest) CacheKey() string {
	return cache.GenerateKey(m.IP(), m.ClusterName(), m.ResourceType(), m.ResourceKind())
}

func (m *MetricsNodeRequest) ClusterName() string {
	return m.Cluster
}

func (m *MetricsNodeRequest) UserID() string {
	return m.UserId
}

func (m *MetricsNodeRequest) OrgID() string {
	return m.OrgId
}

func (m *MetricsNodeRequest) ResourceType() string {
	return m.Type
}

func (m *MetricsNodeRequest) ResourceKind() string {
	return m.Kind
}

func (m *MetricsNodeRequest) IP() string {
	return m.Ip
}

type MetricsData struct {
	// if qurey pod resource, used means usedPercent. request and total are useless.
	Used       float64 `json:"used"`
	Unallocate float64 `json:"unallocate"`
	Left       float64 `json:"left"`
}
