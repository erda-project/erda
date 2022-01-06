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
	Operator        user.ID `gorm:"not null;"`
	ProjectId       uint64
	ProjectName     string
	ApplicationId   int64
	ApplicationName string
	Status          string
	Params          string
	IsOutdated      uint16
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (DeploymentOrder) TableName() string {
	return orderTableName
}

func (db *DBClient) ListDeploymentOrder(projectId uint64, pageInfo *apistructs.PageInfo) (int, []DeploymentOrder, error) {
	cursor := db.Where("project_id = ?", projectId)

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

func (db *DBClient) GetOrderCountByProject(projectId uint64, tp string) (int64, error) {
	if tp == apistructs.TypePipeline {
		return 0, fmt.Errorf("pipeline type doesn't need to count")
	}

	var count int64

	if err := db.Model(&DeploymentOrder{}).Where("project_id = ? and type = ?", projectId, tp).Count(&count).Error; err != nil {
		return 0, errors.Wrapf(err, "failed to count, project: %d, rg: %s", projectId, tp)
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
