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

	"github.com/erda-project/erda/modules/core-services/model"
)

//CreateOrgClusterRelation 创建企业集群关系
func (client *DBClient) CreateOrgClusterRelation(relation *model.OrgClusterRelation) error {
	return client.Create(relation).Error
}

// DeleteOrgClusterRelationByCluster 根据 clusterName 删除企业集群关联关系
func (client *DBClient) DeleteOrgClusterRelationByCluster(clusterName string) error {
	return client.Where("cluster_name = ?", clusterName).Delete(&model.OrgClusterRelation{}).Error
}

// DeleteOrgClusterRelationByClusterAndOrg 根据 clusterName、orgID 删除企业集群关联关系
func (client *DBClient) DeleteOrgClusterRelationByClusterAndOrg(clusterName string, orgID int64) error {
	return client.Where("cluster_name = ?", clusterName).Where("org_id = ?", orgID).Delete(&model.OrgClusterRelation{}).Error
}

// GetOrgClusterRelationByOrgAndCluster 获取企业集群关系
func (client *DBClient) GetOrgClusterRelationByOrgAndCluster(orgID, clusterID int64) (*model.OrgClusterRelation, error) {
	var relation model.OrgClusterRelation
	if err := client.Where("org_id = ?", orgID).
		Where("cluster_id = ?", clusterID).Find(&relation).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &relation, nil
}

// GetOrgClusterRelationsByOrg 根据 orgID 获取企业对应集群关系
func (client *DBClient) GetOrgClusterRelationsByOrg(orgID int64) ([]model.OrgClusterRelation, error) {
	var relations []model.OrgClusterRelation
	if err := client.Where("org_id = ?", orgID).Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

// ListAllOrgClusterRelations 获取所有企业对应集群关系
func (client *DBClient) ListAllOrgClusterRelations() ([]model.OrgClusterRelation, error) {
	var relations []model.OrgClusterRelation
	if err := client.Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}
