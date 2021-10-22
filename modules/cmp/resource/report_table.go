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
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

type ReportTable struct {
	bdl *bundle.Bundle
	cmp interface {
		ListSteveResource(ctx context.Context, req *apistructs.SteveRequest) ([]types.APIObject, error)
		GetNamespacesResources(ctx context.Context, nReq *pb.GetNamespacesResourcesRequest) (*pb.GetNamespacesResourcesResponse, error)
	}
}

func NewReportTable(opt ...ReportTableOption) *ReportTable {
	var rt ReportTable
	for _, f := range opt {
		f(&rt)
	}
	return &rt
}

func (rt *ReportTable) GetResourceOverviewReport(ctx context.Context, orgID int64, clusterNames []string,
	cpuPerNode, memPerNode uint64) (*apistructs.ResourceOverviewReportData, error) {
	// 1) 查找所有 namespaces
	var namespacesM = make(map[string][]string)
	orgIDStr := strconv.FormatInt(orgID, 10)
	for _, clusterName := range clusterNames {
		resources, err := rt.cmp.ListSteveResource(ctx, &apistructs.SteveRequest{
			NoAuthentication: true,
			UserID:           "",
			OrgID:            orgIDStr,
			Type:             apistructs.K8SNamespace,
			ClusterName:      clusterName,
			Name:             "",
			Namespace:        "",
			LabelSelector:    nil,
			FieldSelector:    nil,
			Obj:              nil,
		})
		if err != nil {
			err = errors.Wrap(err, "failed to ListSteveResource")
			logrus.WithError(err).Warnln()
		}
		namespacesM[clusterName] = nil
		for _, resource := range resources {
			namespace := resource.Data().String("metadata", "name")
			namespacesM[clusterName] = append(namespacesM[clusterName], namespace)
		}
	}

	// 2) 调用 core-services bundle，根据 namespaces 查找各 namespaces 的归属
	projectsNamespaces, err := rt.bdl.FetchNamespacesBelongsTo(orgID, namespacesM)
	if err != nil {
		err = errors.Wrap(err, "failed to FetchNamespacesBelongsTo")
		logrus.WithError(err).Errorln()
		return nil, err
	}

	// 3) 查找所有 namespace 下的 request 情况
	var getNamespacesResourcesReq pb.GetNamespacesResourcesRequest
	for clusterName, namespaceList := range namespacesM {
		for _, namespace := range namespaceList {
			getNamespacesResourcesReq.Namespaces = append(getNamespacesResourcesReq.Namespaces, &pb.ClusterNamespacePair{
				ClusterName: clusterName,
				Namespace:   namespace,
			})
		}
	}
	resources, err := rt.cmp.GetNamespacesResources(ctx, &getNamespacesResourcesReq)
	if err != nil {
		err = errors.Wrap(err, "failed to GetNamespacesResources")
		logrus.WithError(err).WithField("request", getNamespacesResourcesReq).Errorln()
		return nil, err
	}

	// 4) request 归属到项目，归属不到项目的，算作共享资源
	var (
		sharedResource [2]uint64
	)
	for _, clusterItem := range resources.List {
		for _, namespaceItem := range clusterItem.List {
			var belongsToProject = false
			for _, projectItem := range projectsNamespaces.List {
				if projectItem.Has(clusterItem.GetClusterName(), namespaceItem.GetNamespace()) {
					belongsToProject = true
					projectItem.AddResource(namespaceItem.GetCpuRequest(), namespaceItem.GetMemRequest())
					break
				}
			}
			if !belongsToProject {
				sharedResource[0] += namespaceItem.GetCpuRequest()
				sharedResource[1] += namespaceItem.GetMemRequest()
			}
		}
	}

	var data apistructs.ResourceOverviewReportData
	for _, projectItem := range projectsNamespaces.List {
		item := apistructs.ResourceOverviewReportDataItem{
			ProjectID:          int64(projectItem.ProjectID),
			ProjectName:        projectItem.ProjectName,
			ProjectDisplayName: projectItem.ProjectDisplayName,
			OwnerUserID:        int64(projectItem.OwnerUserID),
			OwnerUserName:      projectItem.OwnerUserName,
			OwnerUserNickName:  projectItem.OwnerUserNickname,
			CPUQuota:           float64(projectItem.CPUQuota),
			CPUWaterLevel:      float64(projectItem.GetCPUReqeust()) / float64(projectItem.CPUQuota),
			MemQuota:           float64(projectItem.MemQuota),
			MemWaterLevel:      float64(projectItem.GetMemRequest()) / float64(projectItem.MemQuota),
			Nodes:              0,
		}
		item.Nodes = float64(item.CPUQuota) / float64(cpuPerNode)
		if nodes := float64((item.MemQuota)) / float64((memPerNode)); nodes > item.Nodes {
			item.Nodes = nodes
		}
		data.List = append(data.List, &item)
	}

	return &data, nil
}

type ReportTableOption func(table *ReportTable)

func ReportTableWithBundle(bdl *bundle.Bundle) ReportTableOption {
	return func(table *ReportTable) {
		table.bdl = bdl
	}
}

func ReportTableWithCMP(cmp interface {
	ListSteveResource(ctx context.Context, req *apistructs.SteveRequest) ([]types.APIObject, error)
	GetNamespacesResources(ctx context.Context, nReq *pb.GetNamespacesResourcesRequest) (*pb.GetNamespacesResourcesResponse, error)
}) ReportTableOption {
	return func(table *ReportTable) {
		table.cmp = cmp
	}
}
