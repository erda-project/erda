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

package resource

import (
	"fmt"
	"strconv"

	"github.com/rancher/apiserver/pkg/types"

	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
)

const (
	G     = 1 << 30
	mCore = 1000
)

type GaugeData struct {
	Split []float64 `json:"split"`
	Title string    `json:"title"`
	Value []float64 `json:"value"`
	Name  string    `json:"name"`
}

func (r *Resource) GetGauge(ordId string, userID string, request *apistructs.GaugeRequest) (data map[string]*GaugeData, err error) {
	resp, err := r.GetQuotaResource(ordId, userID, request.ClusterName, nil, nil)
	if err != nil {
		return nil, err
	}
	data = r.getGauge(request, resp)
	return data, nil
}

func (r *Resource) getGauge(request *apistructs.GaugeRequest, resp *apistructs.ResourceResp) (data map[string]*GaugeData) {
	data = make(map[string]*GaugeData)
	var (
		nodesGauge = &GaugeData{}
		cpuGauge   = &GaugeData{}
		memGauge   = &GaugeData{}
	)
	data["nodes"] = nodesGauge
	data["cpu"] = cpuGauge
	data["memory"] = memGauge

	if resp.MemTotal == 0 || resp.CpuTotal == 0 {
		return nil
	}
	cpuBase := float64(request.CpuPerNode) * mCore
	memBase := float64(request.MemPerNode) * G
	MemRequest := resp.MemRequest
	CpuRequest := resp.CpuRequest
	MemTotal := resp.MemTotal
	CpuTotal := resp.CpuTotal
	MemQuota := resp.MemQuota
	CpuQuota := resp.CpuQuota

	nodesGauge.Title = r.I18n("节点压力表")
	if MemTotal/memBase > CpuTotal/cpuBase {
		nodesGauge.Value = []float64{MemRequest / MemTotal}
		nodesGauge.Name = fmt.Sprintf("%.1f", MemQuota/MemTotal) + r.I18n("G") + fmt.Sprintf("%.1f%%", nodesGauge.Value[0]) + r.I18n("配额已使用")
		nodesGauge.Split = []float64{MemQuota / MemTotal}
	} else {
		nodesGauge.Value = []float64{CpuRequest / CpuTotal}
		nodesGauge.Name = fmt.Sprintf("%.1f", CpuQuota/CpuTotal) + r.I18n("核") + fmt.Sprintf("%.1f%%", nodesGauge.Value[0]) + r.I18n("配额已使用")
		nodesGauge.Split = []float64{CpuQuota / CpuTotal}
	}
	data["nodes"] = nodesGauge

	cpuGauge.Title = r.I18n("CPU压力表")
	cpuGauge.Value = []float64{CpuRequest / CpuTotal}
	cpuGauge.Name = fmt.Sprintf("%.1f", CpuQuota/CpuTotal) + r.I18n("核") + fmt.Sprintf("%.1f%%", nodesGauge.Value[0]) + r.I18n("配额已使用")
	cpuGauge.Split = []float64{CpuQuota / CpuTotal}
	data["cpu"] = cpuGauge

	memGauge.Title = r.I18n("内存压力表")
	memGauge.Value = []float64{MemRequest / MemTotal}
	memGauge.Name = fmt.Sprintf("%.1f", MemQuota/MemTotal) + r.I18n("G") + fmt.Sprintf("%.1f%%", nodesGauge.Value[0]) + r.I18n("配额已使用")
	memGauge.Split = []float64{MemQuota / MemTotal}
	data["memory"] = memGauge
	return
}

func (r *Resource) GetQuotaResource(ordId string, userID string, clusterNames, projectIds, principal []string) (resp *apistructs.ResourceResp, err error) {
	resp = &apistructs.ResourceResp{}
	orgid, err := strconv.ParseUint(ordId, 10, 64)
	if err != nil {
		return
	}
	clusters, err := r.Bdl.ListClusters("", orgid)
	if err != nil {
		return
	}
	queryCluster := make(map[string]bool)
	for _, name := range clusterNames {
		queryCluster[name] = true
	}
	// 1. filter cluster
	names := make([]string, 0)
	for i := 0; i < len(clusters); i++ {
		if queryCluster[clusters[i].Name] {
			names = append(names, clusters[i].Name)
		}
	}
	// 2. query clusterInfo
	greq := &pb.GetClustersResourcesRequest{}
	greq.ClusterNames = names
	resources, err := r.Server.GetClustersResources(r.Ctx, greq)
	if err != nil {
		return
	}
	// 3. sum request and total of each node
	for _, detail := range resources.List {
		for _, node := range detail.Hosts {
			resp.CpuRequest += float64(node.CpuRequest)
			resp.MemRequest += float64(node.MemRequest)
			resp.CpuTotal += float64(node.CpuTotal)
			resp.MemTotal += float64(node.MemTotal)
		}
	}
	// 4. get all quota
	quota, err := r.Bdl.FetchQuotaOnClusters(orgid, names)
	if err != nil {
		return
	}
	quotaMem, quotaCpu := 0.0, 0.0
	quotaCpu += quota.CPUQuota
	quotaMem += quota.MemQuota

	// 5. get not exist quota
	allNamespace := make([]*pb.ClusterNamespacePair, 0)
	clusterNamespaces := make(map[string][]string)
	for _, clusterName := range names {
		sreq := &apistructs.SteveRequest{
			UserID:      userID,
			OrgID:       ordId,
			Type:        apistructs.K8SNamespace,
			ClusterName: clusterName,
		}
		var resource []types.APIObject
		resource, err = r.Server.ListSteveResource(r.Ctx, sreq)
		if err != nil {
			return
		}
		for _, object := range resource {
			allNamespace = append(allNamespace, &pb.ClusterNamespacePair{Namespace: object.Data().String("metadata", "name"), ClusterName: clusterName})
			clusterNamespaces[cluster] = append(clusterNamespaces[cluster], object.Namespace())
		}
	}
	nreq := &apistructs.OrgClustersNamespaceReq{}
	nreq.OrgID = ordId
	nresp, err := r.Bdl.FetchNamespacesBelongsTo(int64(orgid), clusterNamespaces)
	if err != nil {
		return
	}
	involveNamespace := make(map[string]map[string]bool)
	for _, pair := range nresp.List {
		for k, n := range pair.Clusters {
			for _, s := range n {
				involveNamespace[k][s] = true
			}
		}

	}
	irrelevantNamespace := make([]*pb.ClusterNamespacePair, 0)
	for _, namespace := range allNamespace {
		if !involveNamespace[namespace.ClusterName][namespace.Namespace] {
			irrelevantNamespace = append(irrelevantNamespace, namespace)
		}
	}
	req := &pb.GetNamespacesResourcesRequest{}
	req.Namespaces = irrelevantNamespace
	namespacesResources, err := r.Server.GetNamespacesResources(r.Ctx, req)
	if err != nil {
		return
	}

	for _, irr := range namespacesResources.List {
		for _, detail := range irr.List {
			resp.IrrelevantCpuRequest += float64(detail.CpuRequest)
			resp.IrrelevantMemRequest += float64(detail.MemRequest)
		}
	}

	resp.CpuQuota += quotaCpu + resp.IrrelevantCpuRequest
	resp.MemQuota += quotaMem + resp.IrrelevantMemRequest
	resp.MemRequest += resp.IrrelevantMemRequest
	resp.CpuRequest += resp.IrrelevantCpuRequest
	return
}
