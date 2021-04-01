package dao

import (
	"github.com/erda-project/erda/modules/cmdb/model"
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
