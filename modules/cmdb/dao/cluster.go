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

package dao

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/cmdb/model"
)

// CreateCluster 创建集群
func (client *DBClient) CreateCluster(cluster *model.Cluster) error {
	return client.Create(cluster).Error
}

// UpdateCluster 更新集群
func (client *DBClient) UpdateCluster(cluster *model.Cluster) error {
	return client.Save(cluster).Error
}

// DeleteCluster 删除集群
func (client *DBClient) DeleteCluster(clusterName string) error {
	return client.Where("name = ?", clusterName).Delete(&model.Cluster{}).Error
}

// GetCluster 获取集群详情
func (client *DBClient) GetCluster(clusterID int64) (*model.Cluster, error) {
	var cluster model.Cluster
	if err := client.Where("id = ?", clusterID).Find(&cluster).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &cluster, nil
}

// GetClusterByName 根据集群名称获取集群详情
func (client *DBClient) GetClusterByName(clusterName string) (*model.Cluster, error) {
	var cluster model.Cluster
	if err := client.Where("name = ?", clusterName).First(&cluster).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &cluster, nil
}

// ListCluster 获取集群列表
func (client *DBClient) ListCluster() (*[]model.Cluster, error) {
	var clusters []model.Cluster
	if err := client.Find(&clusters).Error; err != nil {
		return nil, err
	}
	return &clusters, nil
}

// ListClusterByOrg 根据 orgID 获取集群信息
func (client *DBClient) ListClusterByOrg(orgID int64) (*[]model.Cluster, error) {
	var clusters []model.Cluster
	if err := client.Where("org_id = ?", orgID).Find(&clusters).Error; err != nil {
		return nil, err
	}
	return &clusters, nil
}

// ListClusterByOrgAndType 根据 orgID, type 获取集群信息
func (client *DBClient) ListClusterByOrgAndType(orgID int64, clusterType string) (*[]model.Cluster, error) {
	var clusters []model.Cluster
	if orgID == 0 && clusterType == "" {
		return &clusters, nil
	} else if clusterType == "" {
		if err := client.Where("org_id = ?", orgID).Find(&clusters).Error; err != nil {
			return nil, err
		}
	} else if orgID == 0 {
		if err := client.Where("type = ?", clusterType).Find(&clusters).Error; err != nil {
			return nil, err
		}
	} else {
		if err := client.Where("org_id = ?", orgID).Where("type = ?", clusterType).Find(&clusters).Error; err != nil {
			return nil, err
		}
	}
	return &clusters, nil
}

// ListClusterByNames 根据集群名称列表获取集群信息
func (client *DBClient) ListClusterByNames(clusterNames []string) (*[]model.Cluster, error) {
	var clusters []model.Cluster
	if err := client.Where("name in (?)", clusterNames).Find(&clusters).Error; err != nil {
		return nil, err
	}
	return &clusters, nil
}

// ListClusterByIDs 根据集群ID列表获取集群信息
func (client *DBClient) ListClusterByIDs(clusterIDs []uint64) (*[]model.Cluster, error) {
	var clusters []model.Cluster
	if err := client.Where("id in (?)", clusterIDs).Find(&clusters).Error; err != nil {
		return nil, err
	}
	return &clusters, nil
}
