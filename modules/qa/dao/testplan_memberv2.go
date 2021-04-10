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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/dbengine"
)

// AutoTestPlanMember 自动测试计划成员表
type AutoTestPlanMember struct {
	dbengine.BaseModel
	TestPlanID uint64                        `json:"testPlanID"`
	Role       apistructs.TestPlanMemberRole `json:"role"`
	UserID     string                        `json:"userID"`
}

func (AutoTestPlanMember) TableName() string {
	return "dice_autotest_plan_members"
}

func (client *DBClient) GetUserAutoTestPlanRole(userID string, testPlanID uint64) (apistructs.TestPlanMemberRole, error) {
	var mem AutoTestPlanMember
	if err := client.DB.First(&mem, AutoTestPlanMember{UserID: userID, TestPlanID: testPlanID}).Error; err != nil {
		return "", err
	}
	return mem.Role, nil
}

func (client *DBClient) CreateAutoTestPlanMember(mem *AutoTestPlanMember) error {
	return client.DB.Create(mem).Error
}

func (client *DBClient) UpdateAutoTestPlanMember(mem *AutoTestPlanMember) error {
	return client.DB.Save(mem).Error
}

func (client *DBClient) BatchCreateAutoTestPlanMembers(members []AutoTestPlanMember) error {
	return client.BulkInsert(members)
}

// OverwriteAutoTestPlanMembers 使用新的成员列表覆盖之前的成员列表
func (client *DBClient) OverwriteAutoTestPlanMembers(testPlanID uint64, members []AutoTestPlanMember) error {
	if testPlanID == 0 {
		return fmt.Errorf("missing testPlanID")
	}
	// 删除原有成员
	if err := client.Where("`test_plan_id` = ?", testPlanID).Delete(&AutoTestPlanMember{}).Error; err != nil {
		return err
	}
	// 插入新成员
	return client.BatchCreateAutoTestPlanMembers(members)
}

// OverwriteAutoTestPlanOwner 使用新的 owner 覆盖之前的 owner
func (client *DBClient) OverwriteAutoTestPlanOwner(testPlanID uint64, ownerID string) error {
	if testPlanID == 0 {
		return fmt.Errorf("missing testPlanID")
	}
	// 删除原有 owner
	if err := client.Where("`test_plan_id` = ?", testPlanID).
		Where("`role` = ?", apistructs.TestPlanMemberRoleOwner).
		Delete(&AutoTestPlanMember{}).Error; err != nil {
		return err
	}
	// 插入新成员
	return client.BatchCreateAutoTestPlanMembers([]AutoTestPlanMember{{
		TestPlanID: testPlanID,
		Role:       apistructs.TestPlanMemberRoleOwner,
		UserID:     ownerID,
	}})
}

// OverwriteAutoTestPlanPartners 使用新的 partner 列表覆盖之前的 partner 列表
func (client *DBClient) OverwriteAutoTestPlanPartners(testPlanID uint64, partnerIDs []string) error {
	if testPlanID == 0 {
		return fmt.Errorf("missing testPlanID")
	}
	// 删除原有 partners
	if err := client.Where("`test_plan_id` = ?", testPlanID).
		Where("`role` = ?", apistructs.TestPlanMemberRolePartner).
		Delete(&AutoTestPlanMember{}).Error; err != nil {
		return err
	}
	// 插入新成员
	var partners []AutoTestPlanMember
	for _, partnerID := range partnerIDs {
		partners = append(partners, AutoTestPlanMember{
			TestPlanID: testPlanID,
			Role:       apistructs.TestPlanMemberRolePartner,
			UserID:     partnerID,
		})
	}
	return client.BatchCreateAutoTestPlanMembers(partners)
}

func (client *DBClient) DeleteAutoTestPlanMemberByPlanID(planID uint64) error {
	return client.Where("test_plan_id = ?", planID).Delete(AutoTestPlanMember{}).Error
}

func (client *DBClient) ListAutoTestPlanOwnersByOwners(owners []string) ([]AutoTestPlanMember, error) {
	var members []AutoTestPlanMember
	if err := client.Where("user_id in (?) and role = ?", owners, apistructs.TestPlanMemberRoleOwner).
		Find(&members).Error; err != nil {
		return nil, err
	}

	return members, nil
}

func (client *DBClient) ListAutoTestPlanMembersByPlanIDs(testPlanIDs []uint64, roles ...apistructs.TestPlanMemberRole) (map[uint64][]AutoTestPlanMember, error) {
	sql := client.Where("`test_plan_id` IN (?)", testPlanIDs)
	if len(roles) > 0 {
		sql = sql.Where("`role` IN (?)", roles)
	}
	sql = sql.Order("`id` ASC")

	var members []AutoTestPlanMember
	if err := sql.Find(&members).Error; err != nil {
		return nil, err
	}

	testPlanMemberMap := make(map[uint64][]AutoTestPlanMember, len(testPlanIDs))
	for _, mem := range members {
		testPlanMemberMap[mem.TestPlanID] = append(testPlanMemberMap[mem.TestPlanID], mem)
	}
	return testPlanMemberMap, nil
}

func (client *DBClient) ListAutoTestPlanOwnersByPlanID(testPlanID uint64) ([]AutoTestPlanMember, error) {
	return client.ListAutoTestPlanMembersByPlanID(testPlanID, apistructs.TestPlanMemberRoleOwner)
}

func (client *DBClient) ListAutoTestPlanMembersByPlanID(testPlanID uint64, roles ...apistructs.TestPlanMemberRole) ([]AutoTestPlanMember, error) {
	sql := client.Where("`test_plan_id` = ?", testPlanID)
	if len(roles) > 0 {
		sql = sql.Where("`role` IN (?)", roles)
	}
	sql = sql.Order("`id` ASC")

	var members []AutoTestPlanMember
	if err := sql.Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}
