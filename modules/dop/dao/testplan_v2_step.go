// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// TestPlanV2Step 测试计划V2步骤
type TestPlanV2Step struct {
	dbengine.BaseModel
	PlanID     uint64
	SceneSetID uint64
	PreID      uint64
}

// TestPlanV2StepJoin 测试计划V2步骤join测试集表
type TestPlanV2StepJoin struct {
	TestPlanV2Step
	SceneSetName string `gorm:"column:name"`
}

// TableName table name
func (TestPlanV2Step) TableName() string {
	return "dice_autotest_plan_step"
}

// Convert2DTO Convert to apistructs
func (tps TestPlanV2StepJoin) Convert2DTO() *apistructs.TestPlanV2Step {
	return &apistructs.TestPlanV2Step{
		SceneSetID:   tps.SceneSetID,
		SceneSetName: tps.SceneSetName,
		PreID:        tps.PreID,
		PlanID:       tps.PlanID,
		ID:           tps.ID,
	}
}

func (client *DBClient) GetTestPlanV2StepByPreID(preID uint64) (*TestPlanV2Step, error) {
	var step TestPlanV2Step
	if err := client.Where("pre_id = ?", preID).Find(&step).Error; err != nil {
		return nil, err
	}
	return &step, nil
}

func (client *DBClient) GetTestPlanV2Step(ID uint64) (*TestPlanV2StepJoin, error) {
	var step TestPlanV2StepJoin
	err := client.Where("id = ?", ID).First(&step).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

// AddTestPlanV2Step Insert a step in the test plan
func (client *DBClient) AddTestPlanV2Step(req *apistructs.TestPlanV2StepAddRequest) error {
	return client.Transaction(func(tx *gorm.DB) error {
		var preStep, nextStep TestPlanV2Step
		newStep := TestPlanV2Step{PreID: req.PreID, SceneSetID: req.SceneSetID, PlanID: req.TestPlanID}
		// Check the pre step is exist
		if req.PreID != 0 && tx.Where("id = ?", req.PreID).First(&preStep).Error != nil {
			return errors.Errorf("the pre step is not found: %d", req.PreID)
		}
		// Find the next step
		if err := tx.Where("pre_id = ?", req.PreID).Where("plan_id = ?", req.TestPlanID).First(&nextStep).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				// Insert to the end or beginning
				return tx.Create(&newStep).Error
			}
			return err
		}
		// Insert new step
		if err := tx.Create(&newStep).Error; err != nil {
			return err
		}
		// Update the order of next step
		nextStep.PreID = newStep.ID
		return tx.Save(&nextStep).Error
	})
}

// DeleteTestPlanV2Step Delete a step in the test plan
func (client *DBClient) DeleteTestPlanV2Step(req *apistructs.TestPlanV2StepDeleteRequest) error {
	return client.Transaction(func(tx *gorm.DB) error {
		var step, nextStep TestPlanV2Step
		// Get the step
		if err := tx.Where("id = ?", req.StepID).First(&step).Error; err != nil {
			return err
		}
		// Get the next step
		if err := tx.Where("pre_id = ?", req.StepID).Where("plan_id = ?", req.TestPlanID).First(&nextStep).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				// Delete the last step
				return tx.Delete(&step).Error
			}
			return err
		}
		// Update next step
		nextStep.PreID = step.PreID
		if err := tx.Save(&nextStep).Error; err != nil {
			return err
		}

		return tx.Delete(&step).Error
	})
}

// UpdateTestPlanV2Step Update a step in the test plan
func (client *DBClient) MoveTestPlanV2Step(req *apistructs.TestPlanV2StepUpdateRequest) error {
	return client.Transaction(func(tx *gorm.DB) error {
		var step, oldNextStep, newNextStep TestPlanV2Step
		// Get the step
		if err := tx.Where("id = ?", req.StepID).First(&step).Error; err != nil {
			return err
		}
		// Get the old next step
		if err := tx.Where("pre_id = ?", req.StepID).Where("plan_id = ?", req.TestPlanID).First(&oldNextStep).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				// the step was the last step
				goto LABEL1
			}
			return err
		}
		// Update oldNextStep
		oldNextStep.PreID = step.PreID
		if err := tx.Save(&oldNextStep).Error; err != nil {
			return err
		}

	LABEL1:
		// Get the new next step
		step.PreID = req.PreID
		if err := tx.Where("pre_id = ?", req.PreID).Where("plan_id = ?", req.TestPlanID).First(&newNextStep).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				// target step is the last step
				goto LABEL2
			}
			return err
		}
		// Update newNextStep
		newNextStep.PreID = req.StepID
		if err := tx.Save(&newNextStep).Error; err != nil {
			return err
		}

	LABEL2:
		return tx.Save(&step).Error
	})
}

func (client *DBClient) UpdateTestPlanV2Step(step TestPlanV2Step) error {
	return client.Save(&step).Error
}

// GetStepByTestPlanID Get steps of test plan
// if needSort is true then return a sorted list
func (client *DBClient) GetStepByTestPlanID(testPlanID uint64, needSort bool) ([]TestPlanV2StepJoin, int64, error) {
	var (
		steps []TestPlanV2StepJoin
		count int64
	)
	if err := client.Table("dice_autotest_plan_step").Select("dice_autotest_plan_step.id, dice_autotest_plan_step.created_at, "+
		"dice_autotest_plan_step.updated_at, dice_autotest_plan_step.plan_id, dice_autotest_plan_step.pre_id, "+
		"dice_autotest_plan_step.scene_set_id, dice_autotest_scene_set.name").
		Joins("left join dice_autotest_scene_set on dice_autotest_plan_step.scene_set_id = dice_autotest_scene_set.id").
		Where("dice_autotest_plan_step.plan_id = ?", testPlanID).Limit(1000).Scan(&steps).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	return steps, count, nil
}

func (client *DBClient) CheckRelatedSceneSet(setId uint64) (bool, error) {
	var res int
	if err := client.Model(&TestPlanV2Step{}).Where("scene_set_id = ?", setId).Count(&res).Error; err != nil {
		return false, err
	}
	return res > 0, nil
}
