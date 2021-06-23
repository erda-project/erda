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
	"github.com/erda-project/erda/modules/core-services/model"
)

// GetActivitiesByRuntime 根据runtimeID获取活动列表
func (client *DBClient) GetActivitiesByRuntime(runtimeID int64, pageNo, pageSize int) (int, []model.Activity, error) {
	var (
		total      int
		activities []model.Activity
	)
	if err := client.Where("type = ?", "R").
		Where("runtime_id = ?", runtimeID).
		Order("created_at DESC").
		Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&activities).Error; err != nil {
		return 0, nil, err
	}

	// 获取总量
	if err := client.Model(&model.Activity{}).
		Where("type = ?", "R").
		Where("runtime_id = ?", runtimeID).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, activities, nil
}
