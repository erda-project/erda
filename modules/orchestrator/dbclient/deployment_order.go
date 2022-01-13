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

package dbclient

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
)

const (
	orderTableName = "erda_deployment_order"
)

type DeploymentOrder struct {
	ID              string `gorm:"size:36"`
	Name            string
	Type            string
	Description     string
	ReleaseId       string
	Operator        user.ID
	ProjectId       uint64
	ProjectName     string
	ApplicationId   int64
	ApplicationName string
	Workspace       string
	Status          string
	Params          string
	IsOutdated      uint16
	CreatedAt       time.Time
	UpdatedAt       time.Time
	StartedAt       time.Time `gorm:"default:'1970-01-01 00:00:00'"`
}

func (DeploymentOrder) TableName() string {
	return orderTableName
}

func (db *DBClient) ListDeploymentOrder(conditions *apistructs.DeploymentOrderListConditions, pageInfo *apistructs.PageInfo) (int, []DeploymentOrder, error) {
	cursor := db.Where("project_id = ? and workspace = ?", conditions.ProjectId, conditions.Workspace)

	// parse query
	if conditions.Query != "" {
		qv := "%" + conditions.Query + "%"

		// parse user range
		type UserIndex struct {
			Id int
		}
		var UserRange []UserIndex
		if err := db.Table("uc_user").Where("username like ? or nickname like ?", qv, qv).
			Select("id").Scan(&UserRange).Error; err != nil {
			return 0, nil, fmt.Errorf("failed to query release info, err: %v", err)
		}

		// parse release range
		type ReleaseIndex struct {
			ReleaseId string
		}
		var ReleaseRange []ReleaseIndex
		if err := db.Table("dice_release").Where("project_id = ? and (release_id like ? or version like ?)", conditions.ProjectId, qv, qv).
			Select("release_id").Scan(&ReleaseRange).Error; err != nil {
			return 0, nil, fmt.Errorf("failed to query user info, err: %v", err)
		}

		// add query condition
		var (
			uRet = make([]string, 0)
			rRet = make([]string, 0)
		)
		for _, i := range UserRange {
			uRet = append(uRet, strconv.Itoa(i.Id))
		}
		for _, i := range ReleaseRange {
			rRet = append(rRet, i.ReleaseId)
		}

		cursor = cursor.Where("release_id in (?) or operator in (?)", rRet, uRet)
	}

	var (
		total  int
		orders = make([]DeploymentOrder, 0)
	)

	if err := cursor.Order("created_at desc").Offset(pageInfo.GetOffset()).Limit(pageInfo.GetLimit()).Find(&orders).
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return 0, nil, fmt.Errorf("failed to list deployment order, projectId: %d, err: %v", pageInfo, err)
	}

	return total, orders, nil
}

func (db *DBClient) GetOrderCountByProject(tp string, projectId uint64, releaseId string) (int64, error) {
	if tp == apistructs.TypePipeline {
		return 0, fmt.Errorf("pipeline type doesn't need to count")
	}

	var count int64

	if err := db.Model(&DeploymentOrder{}).Where("project_id = ? and release_id = ? and type=?", projectId, releaseId, tp).
		Count(&count).Error; err != nil {
		return 0, errors.Wrapf(err, "failed to count, project: %d, release id: %s, type: %s", projectId, releaseId, tp)
	}

	return count, nil
}

func (db *DBClient) GetDeploymentOrder(id string) (*DeploymentOrder, error) {
	var deploymentOrder DeploymentOrder
	if err := db.
		Where("id = ?", id).
		Find(&deploymentOrder).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get deployment order %s", id)
	}
	return &deploymentOrder, nil
}

func (db *DBClient) UpdateDeploymentOrder(deploymentOrder *DeploymentOrder) error {
	if err := db.Save(deploymentOrder).Error; err != nil {
		return errors.Wrapf(err, "failed to update deployment order, id: %v.",
			deploymentOrder.ID)
	}
	return nil
}
