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
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
)

// ListEdgeSite 获取边缘站点列表
func (c *DBClient) ListEdgeSite(param *apistructs.EdgeSiteListPageRequest) (int, *[]EdgeSite, error) {
	var (
		total     int
		edgeSites []EdgeSite
		db        = c.Model(&EdgeSite{})
	)

	if param.OrgID < 0 && param.ClusterID < 0 {
		return 0, nil, fmt.Errorf("illegal orgin id and cluster id")
	}

	if param.OrgID > 0 && param.ClusterID > 0 {
		db = db.Where("org_id = ? and cluster_id = ?", param.OrgID, param.ClusterID)
	} else if param.OrgID > 0 && param.ClusterID <= 0 {
		db = db.Where("org_id = ?", param.OrgID)
	} else if param.ClusterID > 0 && param.OrgID <= 0 {
		db = db.Where("cluster_id = ?", param.ClusterID)
	}

	if param.Search != "" {
		db = db.Where("locate(?, name)", param.Search)
	}

	db = db.Order("id")

	if param.NotPaging {
		if err := db.Find(&edgeSites).Error; err != nil {
			return 0, nil, err
		}
	} else {
		if err := db.Offset((param.PageNo - 1) * param.PageSize).Limit(param.PageSize).
			Find(&edgeSites).Error; err != nil {
			return 0, nil, err
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, &edgeSites, nil
}

// GetEdgeSite 获取边缘站点详情
func (c *DBClient) GetEdgeSite(edgeSiteID int64) (*EdgeSite, error) {
	var (
		edgeSite EdgeSite
	)
	if err := c.Where("id = ?", edgeSiteID).Find(&edgeSite).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &edgeSite, nil
}

// CreateEdgeSite 创建边缘站点
func (c *DBClient) CreateEdgeSite(edgeSite *EdgeSite) error {
	return c.Create(edgeSite).Error
}

// UpdateEdgeSite 更新边缘站点
func (c *DBClient) UpdateEdgeSite(edgeSite *EdgeSite) error {
	return c.Save(edgeSite).Error
}

// DeleteEdgeSite 删除边缘站点
func (c *DBClient) DeleteEdgeSite(edgeSiteID int64) error {
	return c.Where("id = ?", edgeSiteID).Delete(&EdgeSite{}).Error
}
