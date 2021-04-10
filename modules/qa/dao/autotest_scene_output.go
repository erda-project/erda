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
	"github.com/erda-project/erda/pkg/dbengine"
)

type AutoTestSceneOutput struct {
	dbengine.BaseModel
	Name        string `gorm:"name"`
	Value       string `gorm:"value"`       // 值表达式
	Description string `gorm:"description"` // 描述
	SceneID     uint64 `gorm:"scene_id"`    // 场景ID
	SpaceID     uint64 `gorm:"space_id"`    // 所属测试空间ID
	CreatorID   string `gorm:"creator_id"`
	UpdaterID   string `gorm:"updater_id"`
}

func (AutoTestSceneOutput) TableName() string {
	return "dice_autotest_scene_output"
}

func (db *DBClient) CreateAutoTestSceneOutput(output *AutoTestSceneOutput) error {
	return db.Create(output).Error
}

func (db *DBClient) CreateAutoTestSceneOutputs(output []AutoTestSceneOutput) error {
	return db.BulkInsert(output)
}

func (db *DBClient) UpdateAutotestSceneOutput(output *AutoTestSceneOutput) error {
	return db.Save(output).Error
}

func (db *DBClient) DeleteAutoTestSceneOutput(id uint64) error {
	return db.Where("id = ?", id).Delete(AutoTestSceneOutput{}).Error
}

func (db *DBClient) GetAutoTestSceneOutput(id uint64) (*AutoTestSceneOutput, error) {
	var output *AutoTestSceneOutput
	err := db.Where("id = ?", id).Find(&output).Error
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (db *DBClient) ListAutoTestSceneOutput(sceneID uint64) ([]AutoTestSceneOutput, error) {
	var outputs []AutoTestSceneOutput
	err := db.Where("scene_id = ?", sceneID).Find(&outputs).Error
	if err != nil {
		return nil, err
	}
	return outputs, nil
}

func (db *DBClient) ListAutoTestSceneOutputByScenes(sceneID []uint64) ([]AutoTestSceneOutput, error) {
	var outputs []AutoTestSceneOutput
	err := db.Where("scene_id in (?)", sceneID).Find(&outputs).Error
	if err != nil {
		return nil, err
	}
	return outputs, nil
}
