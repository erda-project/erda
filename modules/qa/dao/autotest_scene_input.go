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

type AutoTestSceneInput struct {
	dbengine.BaseModel
	Name        string `gorm:"name"`
	Value       string `gorm:"value"`       // 默认值
	Temp        string `gorm:"temp"`        // 当前值
	Description string `gorm:"description"` // 描述
	SceneID     uint64 `gorm:"scene_id"`    // 场景ID
	SpaceID     uint64 `gorm:"space_id"`    // 所属测试空间ID
	CreatorID   string `gorm:"creator_id"`
	UpdaterID   string `gorm:"updater_id"`
}

func (AutoTestSceneInput) TableName() string {
	return "dice_autotest_scene_input"
}

func (db *DBClient) CreateAutoTestSceneInput(input *AutoTestSceneInput) error {
	return db.Create(input).Error

}

func (db *DBClient) CreateAutoTestSceneInputs(input []AutoTestSceneInput) error {
	return db.BulkInsert(input)

}

func (db *DBClient) UpdateAutotestSceneInput(input *AutoTestSceneInput) error {
	return db.Save(input).Error
}

func (db *DBClient) DeleteAutoTestSceneInput(id uint64) error {
	return db.Where("id = ?", id).Delete(AutoTestSceneInput{}).Error
}

func (db *DBClient) GetAutoTestSceneInput(id uint64) (*AutoTestSceneInput, error) {
	var input AutoTestSceneInput
	err := db.Where("scene_id = ?", id).Find(&input).Error
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (db *DBClient) ListAutoTestSceneInput(sceneID uint64) ([]AutoTestSceneInput, error) {
	var inputs []AutoTestSceneInput
	err := db.Where("scene_id = ?", sceneID).Find(&inputs).Error
	if err != nil {
		return nil, err
	}
	return inputs, nil
}

func (db *DBClient) ListAutoTestSceneInputByScenes(sceneID []uint64) ([]AutoTestSceneInput, error) {
	var inputs []AutoTestSceneInput
	err := db.Where("scene_id in (?)", sceneID).Find(&inputs).Error
	if err != nil {
		return nil, err
	}
	return inputs, nil
}
