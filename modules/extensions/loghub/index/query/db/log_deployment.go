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

package db

import (
	"strconv"

	"github.com/jinzhu/gorm"
)

// LogDeploymentTable .
const LogDeploymentTable = "sp_log_deployment"

// LogDeployment .
type LogDeployment struct {
	ID           int64  `gorm:"column:id" json:"id"`
	OrgID        string `gorm:"column:org_id" json:"org_id"`
	ClusterName  string `gorm:"column:cluster_name" json:"cluster_name"`
	ClusterType  int    `gorm:"column:cluster_type" json:"cluster_type"`
	ESURL        string `gorm:"column:es_url" json:"es_url"`
	ESConfig     string `gorm:"column:es_config" json:"es_config"`
	CollectorURL string `gorm:"column:collector_url" json:"collector_url"`
	Domain       string `gorm:"column:domain" json:"domain"`
}

// TableName .
func (LogDeployment) TableName() string { return LogDeploymentTable }

// LogDeploymentDB .
type LogDeploymentDB struct {
	*gorm.DB
}

// QueryByOrgID .
func (db *LogDeploymentDB) QueryByOrgID(orgID int64) ([]*LogDeployment, error) {
	var list []*LogDeployment
	if err := db.Table(LogDeploymentTable).
		Where("org_id=?", strconv.FormatInt(orgID, 10)).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// QueryByClusters .
func (db *LogDeploymentDB) QueryByClusters(clusters ...string) ([]*LogDeployment, error) {
	var list []*LogDeployment
	if err := db.Table(LogDeploymentTable).
		Where("cluster_name IN (?)", clusters).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// List .
func (db *LogDeploymentDB) List() ([]*LogDeployment, error) {
	var list []*LogDeployment
	if err := db.Table(LogDeploymentTable).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
