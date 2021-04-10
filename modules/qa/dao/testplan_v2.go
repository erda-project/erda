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
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

// TestPlanV2 测试计划V2
type TestPlanV2 struct {
	dbengine.BaseModel
	Name      string
	Desc      string
	CreatorID string
	UpdaterID string
	ProjectID uint64
	SpaceID   uint64
}

// TableName table name
func (TestPlanV2) TableName() string {
	return "dice_autotest_plan"
}

// Convert2DTO convert DAO to DTO
func (tp *TestPlanV2) Convert2DTO() apistructs.TestPlanV2 {
	return apistructs.TestPlanV2{
		ID:        tp.ID,
		Name:      tp.Name,
		Desc:      tp.Desc,
		ProjectID: tp.ProjectID,
		SpaceID:   tp.SpaceID,
		Creator:   tp.CreatorID,
		Updater:   tp.UpdaterID,
		Steps:     []*apistructs.TestPlanV2Step{},
		CreateAt:  &tp.CreatedAt,
		UpdateAt:  &tp.UpdatedAt,
	}
}

// TestPlanV2Join join dice_autotest_space
type TestPlanV2Join struct {
	TestPlanV2
	SpaceName string
}

// Convert2DTO convert DAO to DTO
func (tp *TestPlanV2Join) Convert2DTO() *apistructs.TestPlanV2 {
	return &apistructs.TestPlanV2{
		ID:        tp.ID,
		Name:      tp.Name,
		Desc:      tp.Desc,
		ProjectID: tp.ProjectID,
		SpaceID:   tp.SpaceID,
		SpaceName: tp.SpaceName,
		Creator:   tp.CreatorID,
		Updater:   tp.UpdaterID,
		Steps:     []*apistructs.TestPlanV2Step{},
	}
}

// CreateTestPlanV2 Create test plan
func (client *DBClient) CreateTestPlanV2(testPlan *TestPlanV2) error {
	return client.Create(testPlan).Error
}

// GetTestPlanV2ByID Get test plan by id
func (client *DBClient) GetTestPlanV2ByID(testPlanID uint64) (*TestPlanV2, error) {
	var testPlan TestPlanV2
	if err := client.Model(&TestPlanV2{}).Where("id = ?", testPlanID).First(&testPlan).Error; err != nil {
		return nil, err
	}

	return &testPlan, nil
}

// DeleteTestPlanV2ByID Delete test plan and his all steps by ID
func (client *DBClient) DeleteTestPlanV2ByID(testPlanID uint64) error {
	return client.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("plan_id = ?", testPlanID).Delete(TestPlanV2Step{}).Error; err != nil {
			return err
		}

		return tx.Where("id = ?", testPlanID).Delete(TestPlanV2{}).Error
	})
}

// UpdateTestPlanV2 Update test plan
func (client *DBClient) UpdateTestPlanV2(testPlanID uint64, fields map[string]interface{}) error {
	tp := TestPlanV2{}
	tp.ID = testPlanID

	return client.Model(&tp).Updates(fields).Error
}

// PagingTestPlanV2 Page query testplan
func (client *DBClient) PagingTestPlanV2(req *apistructs.TestPlanV2PagingRequest) (int, []*apistructs.TestPlanV2, []string, error) {
	var (
		testPlanJoins []TestPlanV2Join
		total         int
	)
	db := client.Table("dice_autotest_plan").Select("dice_autotest_plan.id, dice_autotest_plan.created_at, "+
		"dice_autotest_plan.updated_at, dice_autotest_plan.name, dice_autotest_plan.desc, dice_autotest_plan.creator_id, "+
		"dice_autotest_plan.updater_id, "+"dice_autotest_plan.project_id, dice_autotest_plan.space_id, "+
		"dice_autotest_space.name as space_name").
		Joins("inner join dice_autotest_space on dice_autotest_plan.space_id = dice_autotest_space.id").
		Where("dice_autotest_plan.project_id = ?", req.ProjectID)

	if req.Name != "" {
		db = db.Where("dice_autotest_plan.name LIKE ?", "%"+req.Name+"%")
	}
	if req.Creator != "" {
		db = db.Where("dice_autotest_plan.creator_id = ?", req.Creator)
	}
	if req.Updater != "" {
		db = db.Where("dice_autotest_plan.updater_id = ?", req.Updater)
	}
	if req.SpaceID != 0 {
		db = db.Where("dice_autotest_plan.space_id = ?", req.SpaceID)
	}
	if len(req.IDs) != 0 {
		db = db.Where("dice_autotest_plan.id in (?)", req.IDs)
	}

	if err := db.Order("created_at DESC").Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&testPlanJoins).Offset(0).Limit(-1).
		Count(&total).Error; err != nil {
		return 0, nil, nil, err
	}

	var testPlanIDs []uint64
	result := make([]*apistructs.TestPlanV2, 0, total)
	for _, v := range testPlanJoins {
		testPlanIDs = append(testPlanIDs, v.ID)
		result = append(result, v.Convert2DTO())
	}

	// get owners
	testPlanMember, err := client.ListAutoTestPlanMembersByPlanIDs(testPlanIDs, apistructs.TestPlanMemberRoleOwner)
	if err != nil {
		return 0, nil, nil, err
	}

	// set owner in test plan
	var userIDs []string
	for _, v := range result {
		if _, ok := testPlanMember[v.ID]; !ok {
			continue
		}
		for _, m := range testPlanMember[v.ID] {
			v.Owners = append(v.Owners, m.UserID)
		}
		v.Owners = strutil.DedupSlice(v.Owners)
		userIDs = append(userIDs, v.Owners...)
		userIDs = append(userIDs, v.Updater, v.Creator)
	}
	userIDs = strutil.DedupSlice(userIDs)

	return total, result, userIDs, nil
}

// CheckTestPlanV2NameExist Check if the name of the test plan is repeated
func (client *DBClient) CheckTestPlanV2NameExist(name string) error {
	var count int64
	if err := client.Model(&TestPlanV2{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.Errorf("test plan name is existed: %s", name)
	}

	return nil
}
