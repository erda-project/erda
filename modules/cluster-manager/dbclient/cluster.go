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

package dbclient

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/cluster-manager/model"
)

// CreateCluster create cluster
func (client *DBClient) CreateCluster(cluster *model.Cluster) error {
	return client.Create(cluster).Error
}

// UpdateCluster update cluster
func (client *DBClient) UpdateCluster(cluster *model.Cluster) error {
	return client.Save(cluster).Error
}

// DeleteCluster delete cluster
func (client *DBClient) DeleteCluster(clusterName string) error {
	return client.Where("name = ?", clusterName).Delete(&model.Cluster{}).Error
}

// GetClusterByName get cluster by name
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

// GetClusterByID get cluster by name
func (client *DBClient) GetClusterByID(clusterID int64) (*model.Cluster, error) {
	var cluster model.Cluster
	if err := client.Where("id = ?", clusterID).First(&cluster).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &cluster, nil
}

// ListCluster list cluster
func (client *DBClient) ListCluster() (*[]model.Cluster, error) {
	var clusters []model.Cluster
	if err := client.Find(&clusters).Error; err != nil {
		return nil, err
	}
	return &clusters, nil
}

// ListClusterByType list cluster filter by cluster type
func (client *DBClient) ListClusterByType(clusterType string) (*[]model.Cluster, error) {
	var clusters []model.Cluster
	if clusterType != "" {
		if err := client.Where("type = ?", clusterType).Find(&clusters).Error; err != nil {
			return nil, err
		}
	}
	return &clusters, nil
}
