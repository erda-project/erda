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

package tasks

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/dbclient"
)

type DailyQuotaCollector struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
	cmp interface {
		ListSteveResource(ctx context.Context, req *apistructs.SteveRequest) ([]types.APIObject, error)
		GetNamespacesResources(ctx context.Context, nReq *pb.GetNamespacesResourcesRequest) (*pb.GetNamespacesResourcesResponse, error)
		GetClustersResources(ctx context.Context, cReq *pb.GetClustersResourcesRequest) (*pb.GetClusterResourcesResponse, error)
		GetAllClusters() []string
	}
}

func NewDailyQuotaCollector(opt ...DailyQuotaCollectorOption) *DailyQuotaCollector {
	var d DailyQuotaCollector
	for _, f := range opt {
		f(&d)
	}
	return &d
}

func (d *DailyQuotaCollector) Task() (bool, error) {
	// 1) 查出所有的 clusters
	clusterNames := d.cmp.GetAllClusters()

	// 2) 查出所有的 namespace
	var namespacesM = make(map[string][]string)
	for _, clusterName := range clusterNames {
		resources, err := d.cmp.ListSteveResource(context.Background(), &apistructs.SteveRequest{
			NoAuthentication: true,
			UserID:           "",
			OrgID:            "",
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

	if err := d.collectProjectDaily(namespacesM); err != nil {
		err = errors.Wrap(err, "failed to collectProjectDaily")
		logrus.WithError(err).WithField("namespaces", namespacesM).Errorln()
	}
	if err := d.collecteClusterDaily(clusterNames); err != nil {
		err = errors.Wrap(err, "failed to collecteClusterDaily")
		logrus.WithError(err).WithField("clusters", clusterNames).Errorln()
	}

	return false, nil
}

func (d *DailyQuotaCollector) collectProjectDaily(namespacesM map[string][]string) error {
	// 3) 拿 namespaces 调用 core-services 反查 namespace 的项目归属
	// 该接口查询了 namespace 项目归属，项目 quota，项目 request
	projectsNamespaces, err := d.bdl.FetchNamespacesBelongsTo(0, namespacesM)
	if err != nil {
		err = errors.Wrap(err, "failed to FetchNamespacesBelongsTo")
		logrus.WithError(err).WithField("namespaces", namespacesM).Errorln()
		return err
	}

	for _, item := range projectsNamespaces.List {
		record := apistructs.ProjectResourceDailyModel{
			ID:          0,
			ProjectID:   uint64(item.ProjectID),
			ProjectName: item.ProjectName,
			CPUQuota:    item.CPUQuota,
			CPURequest:  item.GetCPUReqeust(),
			MemQuota:    item.MemQuota,
			MemRequest:  item.GetMemRequest(),
		}
		var existsRecord apistructs.ProjectResourceDailyModel
		err := d.db.Where("updated_at > ? and updated_at < ?",
			time.Now().Format("2006-01-02 00:00:00"),
			time.Now().Add(time.Hour*24).Format("2006-01-02 00:00:00")).
			First(&existsRecord, map[string]interface{}{"project_id": item.ProjectID}).
			Error
		switch {
		case err == nil:
			record.ID = existsRecord.ID
			if err = d.db.Save(&record).Error; err != nil {
				logrus.WithError(err).Errorln("failed to save project resource daily record")
			}
		case gorm.IsRecordNotFoundError(err):
			if err = d.db.Create(&record).Error; err != nil {
				logrus.WithError(err).Errorln("failed to create project resource daily record")
			}
		default:
			err = errors.Wrap(err, "failed to Save or Create project daily record")
			logrus.WithError(err).WithField("project daily record", record).Errorln()
		}
	}

	return nil
}

func (d *DailyQuotaCollector) collecteClusterDaily(clusterNames []string) error {
	// 3) 调用本地接口，查询各 cluster 上的 request
	req := pb.GetClustersResourcesRequest{
		ClusterNames: clusterNames,
	}
	clustersResources, err := d.cmp.GetClustersResources(context.Background(), &req)
	if err != nil {
		err = errors.Wrap(err, "failed to GetClustersResources")
		logrus.WithError(err).WithField("clusterNames", clusterNames).Errorln()
		return err
	}

	// 4) 累计
	var records = make(map[string]*apistructs.ClusterResourceDailyModel)
	for _, cluster := range clustersResources.List {
		record, ok := records[cluster.GetClusterName()]
		if !ok {
			record = &apistructs.ClusterResourceDailyModel{
				ClusterName: cluster.GetClusterName(),
			}
			records[cluster.GetClusterName()] = record
		}

		for _, host := range cluster.Hosts {
			record.CPUTotal += host.GetCpuTotal()
			record.CPURequested += host.GetCpuRequest()
			record.MemTotal += host.GetMemTotal()
			record.MemRequested += host.GetMemRequest()
		}
	}

	// 5) 插入库表
	for clusterName, record := range records {
		var existsRecord apistructs.ClusterResourceDailyModel
		err := d.db.Where("updated_at >= ? and udpated_at < ?",
			time.Now().Format("2006-01-02 00:00:00"),
			time.Now().Add(time.Hour*24).Format("2006-01-02 00:00:00")).
			First(&existsRecord, map[string]interface{}{"cluster_name": clusterName}).
			Error
		switch {
		case err == nil:
			record.ID = existsRecord.ID
			if err = d.db.Save(&record).Error; err != nil {
				logrus.WithError(err).Errorln("failed to save cluster resource daily record")
			}
		case gorm.IsRecordNotFoundError(err):
			if err = d.db.Create(&record).Error; err != nil {
				logrus.WithError(err).Errorln("failed to create cluster resource daily record")
			}
		default:
			err = errors.Wrap(err, "failed to First ClusterResourceDailyModel")
			logrus.WithError(err).Errorln()
		}
	}

	return nil
}

type DailyQuotaCollectorOption func(collector *DailyQuotaCollector)

func DailyQuotaCollectorWithDBClient(db *dbclient.DBClient) DailyQuotaCollectorOption {
	return func(collector *DailyQuotaCollector) {
		collector.db = db
	}
}

func DailyQuotaCollectorWithBundle(bdl *bundle.Bundle) DailyQuotaCollectorOption {
	return func(collector *DailyQuotaCollector) {
		collector.bdl = bdl
	}
}

func DailyQuotaCollectorWithCMPAPI(cmp interface {
	ListSteveResource(ctx context.Context, req *apistructs.SteveRequest) ([]types.APIObject, error)
	GetNamespacesResources(ctx context.Context, nReq *pb.GetNamespacesResourcesRequest) (*pb.GetNamespacesResourcesResponse, error)
	GetClustersResources(ctx context.Context, cReq *pb.GetClustersResourcesRequest) (*pb.GetClusterResourcesResponse, error)
	GetAllClusters() []string
}) DailyQuotaCollectorOption {
	return func(collector *DailyQuotaCollector) {
		collector.cmp = cmp
	}
}
