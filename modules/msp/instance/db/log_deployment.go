/*
 * // Copyright (c) 2021 Terminus, Inc.
 * //
 * // This program is free software: you can use, redistribute, and/or modify
 * // it under the terms of the GNU Affero General Public License, version 3
 * // or later ("AGPL"), as published by the Free Software Foundation.
 * //
 * // This program is distributed in the hope that it will be useful, but WITHOUT
 * // ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * // FITNESS FOR A PARTICULAR PURPOSE.
 * //
 * // You should have received a copy of the GNU Affero General Public License
 * // along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package db

import "github.com/jinzhu/gorm"

type LogDeploymentDB struct {
	*gorm.DB
}

func (db *LogDeploymentDB) GetByClusterName(clusterName string) (*LogDeployment, error) {
	var deployment LogDeployment
	result := db.Table(TableLogDeployment).
		Where("`cluster_name`=?", clusterName).
		Limit(1).
		Find(&deployment)

	if result.RecordNotFound() {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &deployment, nil
}

func (db *LogDeploymentDB) GetByClusterNameAndOrgId(clusterName, orgId string) (*LogDeployment, error) {
	var deployment LogDeployment
	result := db.Table(TableLogDeployment).
		Where("`cluster_name`=?", clusterName).
		Where("`org_id`=?", orgId).
		Limit(1).
		Find(&deployment)

	if result.RecordNotFound() {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &deployment, nil
}
