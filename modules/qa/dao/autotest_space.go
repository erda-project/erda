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
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/dbengine"
)

// AutoTestSpace 测试空间
type AutoTestSpace struct {
	dbengine.BaseModel
	Name        string
	ProjectID   int64
	Description string
	CreatorID   string
	UpdaterID   string
	Status      apistructs.AutoTestSpaceStatus
	// 被复制的源测试空间
	SourceSpaceID *uint64
	// DeletedAt 删除时间
	DeletedAt *time.Time
}

// TableName 表名
func (AutoTestSpace) TableName() string {
	return "dice_autotest_space"
}

// CreateAutoTestSpace 创建测试空间
func (db *DBClient) CreateAutoTestSpace(space *AutoTestSpace) (*AutoTestSpace, error) {
	return space, db.Create(space).Error
}

// ListAutoTestSpaceByProject 项目下获取测试空间列表
func (db *DBClient) ListAutoTestSpaceByProject(projectID int64, pageNo, pageSize int) ([]AutoTestSpace, int, error) {
	var (
		space []AutoTestSpace
		total int
	)
	offset := (pageNo - 1) * pageSize
	if err := db.Where("project_id = ?", projectID).Order("updated_at desc").
		Offset(offset).Limit(pageSize).Find(&space).
		// reset offset & limit before count
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return space, total, nil
}

// GetAutoTestSpace 获取测试空间
func (db *DBClient) GetAutoTestSpace(id uint64) (*AutoTestSpace, error) {
	var space AutoTestSpace

	err := db.Where("id = ?", id).Find(&space).Error
	if err != nil {
		return nil, err
	}
	return &space, nil
}

// UpdateAutoTestSpace 更新测试空间
func (db *DBClient) UpdateAutoTestSpace(space *AutoTestSpace) (*AutoTestSpace, error) {
	err := db.Where("id = ?", space.ID).Save(space).Error
	return space, err
}

// DeleteAutoTestSpace 删除测试空间
func (db *DBClient) DeleteAutoTestSpace(space *AutoTestSpace) (*AutoTestSpace, error) {
	err := db.Delete(space).Error
	return space, err
}

// GetAutotestSpaceByName 通过空间名获取空间
func (db *DBClient) GetAutotestSpaceByName(name string, projectID int64) (*AutoTestSpace, error) {
	var space AutoTestSpace
	err := db.Where("name = ?", name).Where("project_id = ?", projectID).Find(&space).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &space, nil
}

// DeleteAutoTestSpaceRelation 删除测试空间关联
func (db *DBClient) DeleteAutoTestSpaceRelation(spaceID uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除场景集
		var sceneSet *SceneSet
		sceneSet.SpaceID = spaceID
		if err := tx.Delete(sceneSet).Error; err != nil {
			return err
		}
		fmt.Println("delete sceneSet success")
		// 删除场景
		var scene *AutoTestScene
		scene.SpaceID = spaceID
		if err := tx.Delete(scene).Error; err != nil {
			return err
		}
		fmt.Println("delete scene success")
		// 删除场景入参
		var inputs *AutoTestSceneInput
		inputs.SpaceID = spaceID
		if err := tx.Delete(inputs).Error; err != nil {
			return err
		}
		fmt.Println("delete sceneInput success")
		// 删除场景出参
		var outputs *AutoTestSceneOutput
		outputs.SpaceID = spaceID
		if err := tx.Delete(outputs).Error; err != nil {
			return err
		}
		fmt.Println("delete sceneOutout success")
		// 删除场景步骤
		var step *AutoTestSceneStep
		step.SpaceID = spaceID
		if err := tx.Delete(step).Error; err != nil {
			return err
		}
		fmt.Println("delete sceneStep success")
		return nil
	})
}
