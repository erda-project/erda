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
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
)

const (
	ScopePublic = "public"
)

// ListEdgeConfigSetItem 获取边缘配置集配置项
func (c *DBClient) ListEdgeConfigSetItem(param *apistructs.EdgeCfgSetItemListPageRequest) (int, *[]EdgeConfigSetItem, error) {
	var (
		total       int
		cfgSetItems []EdgeConfigSetItem
		db          = c.Model(&EdgeConfigSetItem{})
	)

	if param.ConfigSetID <= 0 {
		return 0, nil, fmt.Errorf("must provide configset id")
	}

	if (param.Scope == ScopePublic && param.SiteID > 0) || param.SiteID < 0 {
		return 0, nil, fmt.Errorf("illegal site id")
	}

	db = db.Where("configset_id = ?", param.ConfigSetID)

	// 如果指定 Scope 则只查询 public，其余指定 siteID 可以按照站点查询
	if param.Scope == ScopePublic {
		db = db.Where("scope = ?", ScopePublic)
	}

	if param.SiteID != 0 {
		db = db.Where("site_id = ?", param.SiteID)
	}

	// TODO: 多字段模糊匹配优化
	if param.Search != "" {
		cfgSet, err := c.GetEdgeConfigSet(param.ConfigSetID)
		if err != nil {
			return 0, nil, fmt.Errorf("get configset error when search configset item: %v", err)
		}
		// Search name
		_, sites, err := c.ListEdgeSite(&apistructs.EdgeSiteListPageRequest{
			OrgID:     cfgSet.OrgID,
			ClusterID: cfgSet.ClusterID,
			NotPaging: true,
			Search:    param.Search,
		})

		if err != nil {
			return 0, nil, fmt.Errorf("get sites error when search configset item: %v", err)
		}

		candidateSites := make([]uint64, 0)

		for _, site := range *sites {
			candidateSites = append(candidateSites, site.ID)
		}

		if strings.Contains(ScopePublic, param.Search) {
			candidateSites = append(candidateSites, 0)
		}

		db = db.Where("locate(?, item_key) or locate(?, item_value) or site_id in (?)",
			param.Search, param.Search, candidateSites,
		)
	}

	db = db.Order("id")

	if param.NotPaging {
		if err := db.Find(&cfgSetItems).Error; err != nil {
			return 0, nil, err
		}
	} else {
		if err := db.Offset((param.PageNo - 1) * param.PageSize).Limit(param.PageSize).
			Find(&cfgSetItems).Error; err != nil {
			return 0, nil, err
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, &cfgSetItems, nil
}

// GetEdgeConfigSetItem 根据 ID 获取配置项信息
func (c *DBClient) GetEdgeConfigSetItem(itemID int64) (*EdgeConfigSetItem, error) {
	var cfgSetItem EdgeConfigSetItem
	if err := c.Where("id = ?", itemID).Find(&cfgSetItem).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &cfgSetItem, nil
}

// GetEdgeConfigSetItemsBySiteID 根据 Site ID 获取配置项信息
func (c *DBClient) GetEdgeConfigSetItemsBySiteID(siteID int64) (*[]EdgeConfigSetItem, error) {
	var (
		cfgSetItem = make([]EdgeConfigSetItem, 0)
	)

	if err := c.Where("site_id = ?", siteID).Find(&cfgSetItem).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &cfgSetItem, nil
}

// CreateEdgeConfigSetItem 创建边缘配置集配置项
func (c *DBClient) CreateEdgeConfigSetItem(cfgSetItem *EdgeConfigSetItem) error {
	return c.Create(cfgSetItem).Error
}

// UpdateEdgeConfigSetItem 更新边缘配置集配置项
func (c *DBClient) UpdateEdgeConfigSetItem(cfgSetItem *EdgeConfigSetItem) error {
	return c.Save(cfgSetItem).Error
}

// DeleteEdgeConfigSetItem 删除边缘配置集配置项
func (c *DBClient) DeleteEdgeConfigSetItem(cfgSetItemID int64) error {
	return c.Where("id = ?", cfgSetItemID).Delete(&EdgeConfigSetItem{}).Error
}

// DeleteEdgeConfigSetItemBySiteID 删除指定站点下的所有配置项
func (c *DBClient) DeleteEdgeConfigSetItemBySiteID(siteID int64) error {
	return c.Where("site_id = ?", siteID).Delete(&EdgeConfigSetItem{}).Error
}

// DeleteEdgeCfgSetItemByCfgID 删除指定配置集ID下的所有配置项
func (c *DBClient) DeleteEdgeCfgSetItemByCfgID(configSetID int64) error {
	return c.Where("configset_id = ?", configSetID).Delete(&EdgeConfigSetItem{}).Error
}
