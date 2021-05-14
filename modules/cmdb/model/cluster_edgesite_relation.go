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

package model

// ClusterEdgeSiteRelation 集群边缘站点关联关系
type ClusterEdgeSiteRelation struct {
	BaseModel
	ClusterID  int64 `gorm:"unique_index:idx_cluster_edgesite_id"`
	EdgeSiteID int64 `gorm:"unique_index:idx_cluster_edgesite_id"`
}

// TableName 设置模型对应数据库表名称
func (ClusterEdgeSiteRelation) TableName() string {
	return "cluster_edgesite_relation"
}
