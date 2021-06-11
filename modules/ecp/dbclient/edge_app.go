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

// ListEdgeApp List edge app by paging
func (c *DBClient) ListEdgeApp(param *apistructs.EdgeAppListPageRequest) (int, *[]EdgeApp, error) {
	var (
		total    int
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if param.OrgID < 0 {
		return 0, nil, fmt.Errorf("illegal orgin id")
	}

	db = db.Where("org_id = ?", param.OrgID).Order("id")

	if err := db.Offset((param.PageNo - 1) * param.PageSize).Limit(param.PageSize).
		Find(&edgeApps).Error; err != nil {
		return 0, nil, err
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, &edgeApps, nil
}

// ListAllEdgeApp List all edge application by orgID
func (c *DBClient) ListAllEdgeApp(orgID int64) (*[]EdgeApp, error) {
	var (
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if orgID < 0 {
		return nil, fmt.Errorf("illegal orgin id")
	}

	if err := db.Where("org_id = ?", orgID).Find(&edgeApps).Error; err != nil {
		return nil, err
	}

	return &edgeApps, nil
}

// ListAllEdgeAppByClusterID List all edge application by orgID and clusterID
func (c *DBClient) ListAllEdgeAppByClusterID(orgID, clusterID int64) (*[]EdgeApp, error) {
	var (
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if orgID < 0 {
		return nil, fmt.Errorf("illegal orgin id")
	}

	if err := db.Where("org_id = ? and cluster_id = ?", orgID, clusterID).Find(&edgeApps).Error; err != nil {
		return nil, err
	}

	return &edgeApps, nil
}

// ListDependsEdgeApps List edge applications which depended
func (c *DBClient) ListDependsEdgeApps(orgID, clusterID int64, appName string) (*[]EdgeApp, error) {
	var (
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if orgID < 0 || clusterID < 0 {
		return nil, fmt.Errorf("illegal orgID or clusterID")
	}

	if err := db.Where("org_id = ? and cluster_id = ? and depend_app like ?",
		orgID, clusterID, "%"+"\""+appName+"\""+"%").Find(&edgeApps).Error; err != nil {
		return nil, err
	}

	return &edgeApps, nil
}

// ListEdgeAppBySiteName List edge application by site name under specified cluster
func (c *DBClient) ListEdgeAppBySiteName(orgID, clusterID int64, siteName string) (*[]EdgeApp, error) {
	var (
		edgeApps []EdgeApp
		db       = c.Model(&EdgeApp{})
	)

	if orgID < 0 || clusterID < 0 {
		return nil, fmt.Errorf("illegal orgin or cluster id")
	}

	if err := db.Where("org_id = ? and cluster_id = ? and edge_sites like ?", orgID,
		clusterID, "%"+"\""+siteName+"\""+"%").Find(&edgeApps).Error; err != nil {
		return nil, err
	}

	return &edgeApps, nil
}

// GetEdgeApp Get edge application by id
func (c *DBClient) GetEdgeApp(edgeAppID int64) (*EdgeApp, error) {
	var edgeApp EdgeApp
	if err := c.Where("id = ?", edgeAppID).Find(&edgeApp).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &edgeApp, nil
}

// GetEdgeAppByName Get edge application by name
func (c *DBClient) GetEdgeAppByName(appName string, orgID int64) (*EdgeApp, error) {
	var edgeApp EdgeApp
	if err := c.Where("name = ? AND org_id = ?", appName, orgID).Find(&edgeApp).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &edgeApp, nil
}

// GetEdgeAppsBySiteName Get edge application
func (c *DBClient) GetEdgeAppsBySiteName(siteName string, clusterID int64) (*[]EdgeApp, error) {
	var edgeApp []EdgeApp
	if err := c.Where("cluster_id = ? and edge_sites like ?", clusterID, "%"+"\""+siteName+"\""+"%").Find(&edgeApp).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &edgeApp, nil
}

// GetEdgeAppByConfigSet Get edge application by configSet name
func (c *DBClient) GetEdgeAppByConfigSet(configSetName string, clusterID int64) (*[]EdgeApp, error) {
	var edgeApp []EdgeApp
	if err := c.Where("cluster_id = ? AND config_set_name = ?", clusterID, configSetName).Find(&edgeApp).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &edgeApp, nil
}

// CreateEdgeApp Create edge application
func (c *DBClient) CreateEdgeApp(edgeApp *EdgeApp) error {
	return c.Create(edgeApp).Error
}

// UpdateEdgeApp Update edge application
func (c *DBClient) UpdateEdgeApp(edgeApp *EdgeApp) error {
	return c.Save(edgeApp).Error
}

// DeleteEdgeApp Delete edge application
func (c *DBClient) DeleteEdgeApp(edgeAppID int64) error {
	return c.Where("id = ?", edgeAppID).Delete(&EdgeApp{}).Error
}
