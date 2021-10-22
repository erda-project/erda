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

import (
	calcu "github.com/erda-project/erda/pkg/resourcecalculator"
)

type GetWorkspaceQuotaRequest struct {
	ProjectID string `json:"projectID"`
	Workspace string `json:"workspace"`
}

type GetWorkspaceQuotaResponse struct {
	Header
	Data WorkspaceQuotaData `json:"data"`
}

type WorkspaceQuotaData struct {
	CPU    int64 `json:"cpu"`
	Memory int64 `json:"memory"`
}

type GetQuotaOnClustersResponse struct {
	ClusterNames []string `json:"clusterNames"`
	// CPUQuota is the total cpu quota on the clusters
	CPUQuota float64 `json:"cpuQuota"`
	cpuQuota uint64
	// MemQuota is hte total mem quota on the clusters
	MemQuota float64 `json:"memQuota"`
	memQuota uint64
	Owners   []*OwnerQuotaOnClusters `json:"owners"`
}

// AccuQuota accumulate cpu and mem quota value
func (q *GetQuotaOnClustersResponse) AccuQuota(cpu, mem uint64) {
	q.cpuQuota += cpu
	q.memQuota += mem
}

func (q GetQuotaOnClustersResponse) ReCalcu() {
	q.CPUQuota = 0
	q.MemQuota = 0
	for _, owner := range q.Owners {
		owner.ReCalcu()
		q.AccuQuota(owner.cpuQuota, owner.memQuota)
	}
	q.CPUQuota = calcu.MillcoreToCore(q.cpuQuota)
	q.MemQuota = calcu.ByteToGibibyte(q.memQuota)
}

type OwnerQuotaOnClusters struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	// CPUQuota is the total cpu quota for the owner on the clusters
	CPUQuota float64 `json:"cpuQuota"`
	cpuQuota uint64
	// MemQuota is the total mem quota for the owner on the clusters
	MemQuota float64 `json:"memQuota"`
	memQuota uint64
	Projects []*ProjectQuotaOnClusters `json:"projects"`
}

// AccuQuota accumulate cpu and mem quota value
func (q *OwnerQuotaOnClusters) AccuQuota(cpu, mem uint64) {
	q.cpuQuota += cpu
	q.memQuota += mem
}

func (q OwnerQuotaOnClusters) ReCalcu() {
	q.CPUQuota = 0
	q.MemQuota = 0
	for _, project := range q.Projects {
		project.ReCalcu()
		q.AccuQuota(project.cpuQuota, project.memQuota)
	}
	q.CPUQuota = calcu.MillcoreToCore(q.cpuQuota)
	q.MemQuota = calcu.ByteToGibibyte(q.memQuota)
}

type ProjectQuotaOnClusters struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	// CPUQuota is the total cpu quota for the project on the clusters
	CPUQuota float64 `json:"cpuQuota"`
	cpuQuota uint64
	// CPUQuota is the total mem quota for the project on the clusters
	MemQuota float64 `json:"memQuota"`
	memQuota uint64
}

// AccuQuota accumulate cpu and mem quota value
func (q *ProjectQuotaOnClusters) AccuQuota(cpu, mem uint64) {
	q.cpuQuota += cpu
	q.memQuota += mem
}

func (q ProjectQuotaOnClusters) ReCalcu() {
	q.CPUQuota = calcu.MillcoreToCore(q.cpuQuota)
	q.MemQuota = calcu.ByteToGibibyte(q.memQuota)
}

type GetProjectsNamesapcesResponseData struct {
	Total uint32               `json:"total"`
	List  []*ProjectNamespaces `json:"list"`
}

func (d *GetProjectsNamesapcesResponseData) GetProjectNamespaces(id uint) (*ProjectNamespaces, bool) {
	for _, p := range d.List {
		if p.ProjectID == id {
			return p, true
		}
	}
	return nil, false
}

type ProjectNamespaces struct {
	ProjectID          uint   `json:"projectID"`
	ProjectName        string `json:"projectName"`
	ProjectDisplayName string `json:"projectDisplayName"`
	OwnerUserID        uint   `json:"ownerUserID"`
	OwnerUserName      string `json:"ownerUserName"`
	OwnerUserNickname  string `json:"ownerUserNickname"`
	CPUQuota           uint64 `json:"cpuQuota"`
	MemQuota           uint64 `json:"memQuota"`
	// Clusters the key is cluster name, the value is the list of namespaces
	Clusters map[string][]string `json:"clusters"`

	cpuRequest uint64
	memRequest uint64
}

func (p *ProjectNamespaces) AddResource(cpu, mem uint64) {
	p.cpuRequest += cpu
	p.memRequest += mem
}

func (p *ProjectNamespaces) GetCPUReqeust() uint64 {
	return p.cpuRequest
}

func (p *ProjectNamespaces) GetMemRequest() uint64 {
	return p.memRequest
}

func (p *ProjectNamespaces) Has(cluster, namespace string) bool {
	namespaces, ok := p.Clusters[cluster]
	if !ok {
		return false
	}
	for _, name := range namespaces {
		if name == namespace {
			return true
		}
	}
	return false
}
