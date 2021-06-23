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
	"github.com/erda-project/erda/apistructs"
)

// Label 标签
type Label struct {
	BaseModel
	Name      string                      // 标签名称
	Type      apistructs.ProjectLabelType // 标签作用类型
	Color     string                      // 标签颜色
	ProjectID uint64                      // 标签所属项目
	Creator   string                      // 创建人
}

// TableName 表名
func (Label) TableName() string {
	return "dice_labels"
}

// CreateLabel 创建 label
func (client *DBClient) CreateLabel(label *Label) error {
	return client.Create(label).Error
}

// UpdateLabel 更新 label
func (client *DBClient) UpdateLabel(label *Label) error {
	return client.Save(label).Error
}

// DeleteLabel 删除标签
func (client *DBClient) DeleteLabel(labelID int64) error {
	return client.Where("id = ?", labelID).Delete(&Label{}).Error
}

// ListLabel 获取标签列表
func (client *DBClient) ListLabel(req *apistructs.ProjectLabelListRequest) (int64, []Label, error) {
	var (
		total  int64
		labels []Label
	)
	cond := Label{
		ProjectID: req.ProjectID,
	}
	if req.Type != "" {
		cond.Type = req.Type
	}
	sql := client.Where(cond)
	if req.Key != "" {
		sql = sql.Where("name LIKE ?", "%"+req.Key+"%")
	}
	if err := sql.Order("updated_at desc").Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).
		Find(&labels).Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, labels, nil
}

// GetLabel 根据标签ID获取标签详情
func (client *DBClient) GetLabel(labelID int64) (*Label, error) {
	var label Label
	if err := client.Where("id = ?", labelID).Find(&label).Error; err != nil {
		return nil, err
	}
	return &label, nil
}

// GetLabelByName 根据 name 获取标签详情
func (client *DBClient) GetLabelByName(projectID uint64, name string) (*Label, error) {
	var label Label
	if err := client.Where("project_id = ?", projectID).Where("name = ?", name).First(&label).Error; err != nil {
		return nil, err
	}
	return &label, nil
}

// ListLabelByNames 根据 name 列表获取标签列表
func (client *DBClient) ListLabelByNames(projectID uint64, names []string) ([]Label, error) {
	var labels []Label
	if err := client.Where("project_id = ?", projectID).Where("name in (?)", names).Find(&labels).Error; err != nil {
		return nil, err
	}
	return labels, nil
}

// GetLabels 根据 labelID 获取标签列表
func (client *DBClient) GetLabels(labelIDs []uint64) ([]Label, error) {
	var labels []Label
	if err := client.Where("id in (?)", labelIDs).Find(&labels).Error; err != nil {
		return nil, err
	}
	return labels, nil
}
