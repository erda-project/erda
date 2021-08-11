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
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type AutoTestScene struct {
	dbengine.BaseModel
	Name        string                 `gorm:"name"`
	Description string                 `gorm:"description"` // 描述
	SpaceID     uint64                 `gorm:"space_id"`    // 场景所属测试空间ID
	SetID       uint64                 `gorm:"set_id"`      // 场景集ID
	PreID       uint64                 `gorm:"pre_id"`      // 排序的前驱ID
	CreatorID   string                 `gorm:"creator_id"`
	UpdaterID   string                 `gorm:"updater_id"`
	Status      apistructs.SceneStatus `gorm:"status"`
	RefSetID    uint64                 `gorm:"ref_set_id"` // 引用场景集ID
}

func (AutoTestScene) TableName() string {
	return "dice_autotest_scene"
}

func (s *AutoTestScene) Convert() apistructs.AutoTestScene {
	return apistructs.AutoTestScene{
		AutoTestSceneParams: apistructs.AutoTestSceneParams{
			ID:        s.ID,
			CreatorID: s.CreatorID,
			UpdaterID: s.UpdaterID,
			SpaceID:   s.SpaceID,
		},
		Name:        s.Name,
		Description: s.Description,
		PreID:       s.PreID,
		SetID:       s.SetID,
		CreateAt:    &s.CreatedAt,
		UpdateAt:    &s.UpdatedAt,
		Status:      s.Status,
		RefSetID:    s.RefSetID,
	}
}

func (db *DBClient) CreateAutotestScene(node *AutoTestScene) error {
	return db.Create(node).Error
}

func (db *DBClient) UpdateAutotestScene(node *AutoTestScene) error {
	return db.Save(node).Error
}

func (db *DBClient) GetAutotestScene(id uint64) (*AutoTestScene, error) {
	var scene AutoTestScene
	err := db.Where("id = ?", id).Find(&scene).Error
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

func (db *DBClient) GetAutotestSceneByPreID(preID uint64) (*AutoTestScene, error) {
	var scene AutoTestScene
	err := db.Where("pre_id = ?", preID).Find(&scene).Error
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

func (db *DBClient) GetAutotestSceneByName(name string, setID uint64) (*AutoTestScene, error) {
	var scene AutoTestScene
	err := db.Where("name = ?", name).Where("set_id = ?", setID).Find(&scene).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &scene, nil
}

func (db *DBClient) GetAutotestSceneFirst(setID uint64) (*AutoTestScene, error) {
	var scene AutoTestScene
	err := db.Where("set_id = ?", setID).Where("pre_id = 0").Find(&scene).Error
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

func (db *DBClient) ListAutotestScene(req apistructs.AutotestSceneRequest) (uint64, []AutoTestScene, error) {
	var (
		scenes []AutoTestScene
		total  int64
	)
	//sql := db.Table("dice_autotest_scene").Where("set_id = ?", req.SetID)
	//err := sql.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&scenes).
	//	Offset(0).Limit(-1).Count(&total).Error
	sql := db.Table("dice_autotest_scene").Where("set_id = ?", req.SetID)
	err := sql.Find(&scenes).Offset(0).Limit(-1).Count(&total).Error
	if err != nil {
		return 0, nil, err
	}
	return uint64(total), scenes, nil
}

// ListAutotestScenes 批量查询场景
func (db *DBClient) ListAutotestScenes(setIDs []uint64) ([]AutoTestScene, error) {
	var (
		scenes []AutoTestScene
	)
	err := db.Table("dice_autotest_scene").Where("set_id in (?)", setIDs).Find(&scenes).Error
	if err != nil {
		return nil, err
	}
	return scenes, nil
}

func (db *DBClient) UpdateAutotestSceneUpdater(sceneID uint64, userID string) error {
	return db.Table("dice_autotest_scene").Where("id = ?", sceneID).Update("updater_id", userID).Error
}

func (db *DBClient) UpdateAutotestSceneUpdateAt(sceneID uint64, time time.Time) error {
	return db.Table("dice_autotest_scene").Where("id = ?", sceneID).Update("updated_at", time).Error
}

func (db *DBClient) DeleteAutoTestScene(id uint64) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		var scene, next AutoTestScene
		// 获取scene
		if err := tx.Where("id = ?", id).Find(&scene).Error; err != nil {
			return err
		}
		// 获取next并更新
		if err := tx.Where("pre_id = ?", id).Find(&next).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				goto LABEL1
			}
			return err
		}
		next.PreID = scene.PreID
		if err := tx.Save(&next).Error; err != nil {
			return err
		}

		defer func() {
			err = checkSamePreID(tx, next.SetID, next.PreID)
			if err != nil {
				err = fmt.Errorf("set_id %v have same pre_id %v, please refresh", next.SetID, next.PreID)
			}
		}()
	LABEL1:
		// 删除该场景的全部关联
		if err := tx.Delete(&scene, "pre_id = ? and id = ?", scene.PreID, scene.ID).Error; err != nil {
			return err
		}
		if err := tx.Where(AutoTestSceneInput{}).Where("scene_id = ?", scene.ID).Delete(AutoTestSceneInput{}).Error; err != nil {
			return err
		}
		if err := tx.Where(AutoTestSceneOutput{}).Where("scene_id = ?", scene.ID).Delete(AutoTestSceneOutput{}).Error; err != nil {
			return err
		}
		if err := tx.Where(AutoTestSceneStep{}).Where("scene_id = ?", scene.ID).Delete(AutoTestSceneStep{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// like linklist change to node index
func (db *DBClient) MoveAutoTestScene(id, newPreID, newSetID uint64) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		var scene, next, oldNext AutoTestScene
		// get scene
		if err := tx.Where("id = ?", id).Find(&scene).Error; err != nil {
			return err
		}
		// get next scene
		if err := tx.Where("pre_id = ?", id).Find(&oldNext).Error; err != nil {
			// not have next scene jump over update next scene
			if gorm.IsRecordNotFoundError(err) {
				goto LABEL1
			}
			return err
		}
		// next scene link this scene pre scene
		oldNext.PreID = scene.PreID
		if err := tx.Save(&oldNext).Error; err != nil {
			return err
		}

		defer func() {
			err = checkSamePreID(tx, oldNext.SetID, oldNext.PreID)
			if err != nil {
				err = fmt.Errorf("set_id %v have same pre_id %v, please refresh", oldNext.SetID, oldNext.PreID)
			}
		}()

	LABEL1:
		// get new next scene pre scene
		if err := tx.Where("pre_id = ?", newPreID).Where("set_id = ?", newSetID).Find(&next).Error; err != nil {
			// not have new next scene jump over update new next scene
			if gorm.IsRecordNotFoundError(err) {
				goto LABEL2
			}
			return err
		}
		// new next scene link this scene
		next.PreID = scene.ID
		if err := tx.Save(&next).Error; err != nil {
			return err
		}

		defer func() {
			err = checkSamePreID(tx, next.SetID, next.PreID)
			if err != nil {
				err = fmt.Errorf("set_id %v have same pre_id %v, please refresh", next.SetID, next.PreID)
			}
		}()

	LABEL2:
		// this scene link new next scene pre scene
		scene.SetID = newSetID
		scene.PreID = newPreID
		if err := tx.Save(&scene).Error; err != nil {
			return err
		}

		defer func() {
			err = checkSamePreID(tx, scene.SetID, scene.PreID)
			if err != nil {
				err = fmt.Errorf("set_id %v have same pre_id %v, please refresh", next.SetID, next.PreID)
			}
		}()

		return nil
	})
}

// check sceneSet linked list not have same pre_id
func checkSamePreID(tx *gorm.DB, setId uint64, preID uint64) error {
	var res int
	if err := tx.Model(&AutoTestScene{}).Where("`set_id` = ? and pre_id = ?", setId, preID).Count(&res).Error; err != nil {
		return err
	}
	if res > 1 {
		return fmt.Errorf("set_id %v have same pre_id %v, please refresh", setId, preID)
	}
	return nil
}

func (db *DBClient) GetAutoTestScenePreByPosition(req apistructs.AutotestSceneRequest) (uint64, uint64, bool, error) {
	var next AutoTestScene
	if req.Position == -1 {
		if err := db.Where("id = ?", req.Target).Find(&next).Error; err != nil {
			return 0, 0, false, err
		}
		// 目标的前一个位置是要移动的场景本身,不需要移动
		if next.PreID == req.ID {
			fmt.Println("ahsdiuhasdiouhas")
			return 0, 0, true, nil
		}
		return next.SetID, next.PreID, false, nil
	}
	if err := db.Where("id = ?", req.Target).Find(&next).Error; err != nil {
		return 0, 0, false, err
	}
	return next.SetID, next.ID, false, nil

}

func (db *DBClient) FindSceneBySetAndName(setId uint64, name string) ([]AutoTestScene, error) {
	var scenes []AutoTestScene
	if err := db.Model(&AutoTestScene{}).Where("`set_id` = ? and name = ?", setId, name).Find(&scenes).Error; err != nil {
		return nil, err
	}
	return scenes, nil
}

func (db *DBClient) CountSceneBySetID(setId uint64) (uint64, error) {
	var res uint64
	if err := db.Model(&AutoTestScene{}).Where("set_id = ?", setId).Count(&res).Error; err != nil {
		return 0, err
	}
	return res, nil
}

func (db *DBClient) CountSceneBySpaceID(spaceID uint64) (uint64, error) {
	var res uint64
	if err := db.Model(&AutoTestScene{}).Where("space_id = ?", spaceID).Count(&res).Error; err != nil {
		return 0, err
	}
	return res, nil
}

func (db *DBClient) Insert(scene *AutoTestScene, id uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if id == 0 {
			return db.Create(&scene).Error
		}
		var next AutoTestScene
		if err := db.Where("pre_id = ?", id).Find(&next).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return db.Create(&scene).Error
			}
			return err
		}
		if err := db.Create(&scene).Error; err != nil {
			return err
		}
		next.PreID = scene.ID
		return db.Save(&next).Error
	})
}

func (db *DBClient) UpdateSceneRefSetID(copyRefs apistructs.AutoTestSceneCopyRef) error {
	return db.Model(&AutoTestScene{}).
		Where("space_id = ?", copyRefs.AfterSpaceID).
		Where("ref_set_id = ?", copyRefs.PreSetID).
		Update(map[string]interface{}{"ref_set_id": copyRefs.AfterSetID}).Error
}
