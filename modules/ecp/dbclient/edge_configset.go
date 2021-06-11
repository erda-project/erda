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

// ListEdgeConfigSet List edge configSet
func (c *DBClient) ListEdgeConfigSet(param *apistructs.EdgeConfigSetListPageRequest) (int, *[]EdgeConfigSet, error) {
	var (
		total          int
		edgeConfigSets []EdgeConfigSet
		db             = c.Model(&EdgeConfigSet{})
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

	db = db.Order("id")

	if param.NotPaging {
		if err := db.Find(&edgeConfigSets).Error; err != nil {
			return 0, nil, err
		}
	} else {
		if err := db.Offset((param.PageNo - 1) * param.PageSize).Limit(param.PageSize).
			Find(&edgeConfigSets).Error; err != nil {
			return 0, nil, err
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, &edgeConfigSets, nil
}

// GetEdgeConfigSet Get edge configSet by configSet id
func (c *DBClient) GetEdgeConfigSet(configSetID int64) (*EdgeConfigSet, error) {
	var edgeConfigSet EdgeConfigSet

	if err := c.Where("id = ?", configSetID).Find(&edgeConfigSet).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	fmt.Println(edgeConfigSet)
	return &edgeConfigSet, nil
}

// CreateEdgeConfigSet Create edge configSet
func (c *DBClient) CreateEdgeConfigSet(edgeConfigSet *EdgeConfigSet) error {
	return c.Create(edgeConfigSet).Error
}

// UpdateEdgeConfigSet Update edge configSet
func (c *DBClient) UpdateEdgeConfigSet(edgeConfigSet *EdgeConfigSet) error {
	return c.Save(edgeConfigSet).Error
}

// DeleteEdgeConfigSet Delete edge configSet
func (c *DBClient) DeleteEdgeConfigSet(edgeConfigSetID int64) error {
	return c.Where("id = ?", edgeConfigSetID).Delete(&EdgeConfigSet{}).Error
}
