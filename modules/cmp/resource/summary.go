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
	"math"
	"strconv"

	"github.com/rancher/apiserver/pkg/types"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
)

const (
	G         = 1 << 30
	MilliCore = 1000
)

type GaugeData struct {
	Split []float64 `json:"split"`
	Title string    `json:"title"`
	Value []float64 `json:"value"`
	Name  string    `json:"name"`
}

func (r *Resource) GetGauge(ordId string, userID string, request *apistructs.GaugeRequest) (data map[string]*GaugeData, err error) {
	logrus.Debug("func GetGauge start")
	defer logrus.Debug("func GetGauge finished")
	resp, err := r.GetQuotaResource(ordId, userID, request.ClusterName)
	if err != nil {
		return nil, err
	}
	data = r.getGauge(request, resp)
	return data, nil
}

func (r *Resource) getGauge(req *apistructs.GaugeRequest, resp *apistructs.ResourceResp) (data map[string]*GaugeData) {
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
	cpuBase := float64(req.CpuPerNode) * MilliCore
	memBase := float64(req.MemPerNode) * G
	MemRequest := resp.MemRequest
	CpuRequest := resp.CpuRequest
	MemTotal := resp.MemTotal
	CpuTotal := resp.CpuTotal
	MemQuota := resp.MemQuota
	CpuQuota := resp.CpuQuota

	nodesGauge.Title = r.I18n("node pressure")
	if MemTotal/memBase > CpuTotal/cpuBase {
		nodesGauge.Value = []float64{MemRequest / MemTotal * 100}
		nodesGauge.Name = fmt.Sprintf("%d", int64(math.Round(MemRequest/G+0.5))) + r.I18n("resourceNodeCount") + fmt.Sprintf("\n%.1f%%", nodesGauge.Value[0]) + r.I18n("quota in use")
		nodesGauge.Split = []float64{MemQuota / MemTotal}
	} else {
		nodesGauge.Value = []float64{CpuRequest / CpuTotal * 100}
		nodesGauge.Name = fmt.Sprintf("%d", int64(math.Round(CpuRequest/MilliCore+0.5))) + r.I18n("resourceNodeCount") + fmt.Sprintf("\n%.1f%%", nodesGauge.Value[0]) + r.I18n("quota in use")
		nodesGauge.Split = []float64{CpuQuota / CpuTotal}
	}
	data["nodes"] = nodesGauge

	cpuGauge.Title = r.I18n("cpu pressure")
	cpuGauge.Value = []float64{CpuRequest / CpuTotal * 100}
	cpuGauge.Name = fmt.Sprintf("%.1f", CpuRequest/MilliCore) + r.I18n("core") + fmt.Sprintf("\n%.1f%%", cpuGauge.Value[0]) + r.I18n("quota in use")
	cpuGauge.Split = []float64{CpuQuota / CpuTotal}
	data["cpu"] = cpuGauge

	memGauge.Title = r.I18n("memory pressure")
	memGauge.Value = []float64{MemRequest / MemTotal * 100}
	memGauge.Name = fmt.Sprintf("%.1f", MemRequest/G) + r.I18n("GB") + fmt.Sprintf("\n%.1f%%", memGauge.Value[0]) + r.I18n("quota in use")
	memGauge.Split = []float64{MemQuota / MemTotal}
	data["memory"] = memGauge
	return
}

func (r *Resource) GetQuotaResource(ordId string, userID string, clusterNames []string) (resp *apistructs.ResourceResp, err error) {
	resp = &apistructs.ResourceResp{}
	orgid, err := strconv.ParseUint(ordId, 10, 64)
	if err != nil {
		return
	}
	logrus.Debug("start list cluster")
	clusters, err := r.Bdl.ListClusters("", orgid)
	logrus.Debug("list cluster finished")
	if err != nil {
		return
	}
	// 1. filter Cluster
	names := r.FilterCluster(clusters, clusterNames)
	if len(names) == 0 {
		return nil, errNoClusterFound
	}
	// 2. query clusterInfo
	greq := &pb.GetClustersResourcesRequest{}
	greq.ClusterNames = names
	logrus.Debug("start get cluster resource from steve")
	resources, err := r.Server.GetClustersResources(r.Ctx, greq)
	logrus.Debug("get cluster resource from steve finished")
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
	logrus.Debug("start fetch quota")
	quota, err := r.Bdl.FetchQuotaOnClusters(orgid, names)
	logrus.Debug("fetch quota finished")
	if err != nil {
		return
	}
	quotaMem, quotaCpu := 0.0, 0.0
	quotaCpu += float64(quota.CPUQuotaMilliValue)
	quotaMem += float64(quota.MemQuotaByte)

	// 5. get not exist quota
	logrus.Debug("start get all namespace")
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
			logrus.Error(err)
			continue
		}
		for _, object := range resource {
			allNamespace = append(allNamespace, &pb.ClusterNamespacePair{Namespace: object.Data().String("metadata", "name"), ClusterName: clusterName})
			clusterNamespaces[Cluster] = append(clusterNamespaces[Cluster], object.Namespace())
		}
	}
	logrus.Debug("get all namespace finished")
	logrus.Debug("start involved namespace")

	nresp, err := r.Bdl.FetchNamespacesBelongsTo()
	logrus.Debug("involved namespace finished")
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
	logrus.Debug("get not involve request")
	namespacesResources, err := r.Server.GetNamespacesResources(r.Ctx, req)
	logrus.Debug("get not involve request finished")
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
	return
}

func (r *Resource) FilterCluster(clusters []apistructs.ClusterInfo, clusterNames []string) []string {
	names := make([]string, 0)
	queryCluster := make(map[string]bool)
	for _, name := range clusterNames {
		queryCluster[name] = true
	}
	if len(queryCluster) == 0 {
		for i := 0; i < len(clusters); i++ {
			names = append(names, clusters[i].Name)
		}
	} else {
		for i := 0; i < len(clusters); i++ {
			if queryCluster[clusters[i].Name] {
				names = append(names, clusters[i].Name)
			}
		}
	}
	return names
}
