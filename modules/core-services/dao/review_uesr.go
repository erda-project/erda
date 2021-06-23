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
	"github.com/erda-project/erda/modules/core-services/model"
)

func (client *DBClient) GetOperatorByTaskID(taskID []int) ([]model.ReviewUser, error) {
	var reviewUsers []model.ReviewUser
	err := client.Table("dice_manual_review_user").Where("task_id in (?)", taskID).Find(&reviewUsers).Error
	if err != nil {
		return nil, err
	}
	return reviewUsers, nil
}

func (client *DBClient) GetTaskIDByOperator(param *apistructs.GetReviewsByUserIdRequest) ([]int64, error) {
	user := client.Table("dice_manual_review_user").Where("operator = ?", param.UserId).Where("org_id = ?", param.OrgId)
	var ids []model.ReviewUser
	err := user.Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	var tasks []int64
	for _, v := range ids {
		tasks = append(tasks, v.TaskId)
	}
	return tasks, nil
}

func (client *DBClient) GetAuthorityByUserId(param *apistructs.GetAuthorityByUserIdRequest) (apistructs.GetAuthorityByUserIdResponse, error) {
	var total int
	var authority apistructs.GetAuthorityByUserIdResponse
	authority.Authority = "NO"
	err := client.Table("dice_manual_review_user").Where("operator = ?", param.Operator).Where("org_id = ?", param.OrgId).Where("task_id = ?", param.TaskId).Count(&total).Error
	if err != nil {
		return authority, err
	}
	if total > 0 {
		authority.Authority = "YES"
	}
	return authority, nil
}
