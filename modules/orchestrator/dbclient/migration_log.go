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

package dbclient

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// MigrationLog migration执行记录表
type MigrationLog struct {
	dbengine.BaseModel
	ProjectID           uint64
	ApplicationID       uint64
	RuntimeID           uint64
	DeploymentID        uint64
	OperatorID          uint64
	Status              string
	AddonInstanceID     string
	AddonInstanceConfig string
}

// TableName 数据库表名
func (MigrationLog) TableName() string {
	return "dice_db_migration_log"
}

// CreateMigrationLog insert migrationLog
func (db *DBClient) CreateMigrationLog(migrationLog *MigrationLog) error {
	return db.Create(migrationLog).Error
}

// UpdateMigrationLog update migrationLog
func (db *DBClient) UpdateMigrationLog(migrationLog *MigrationLog) error {
	return db.Save(migrationLog).Error
}

// GetMigrationLogByDeploymentID 根据 deployID 查询migration信息
func (db *DBClient) GetMigrationLogByDeploymentID(deploymentID uint64) (*MigrationLog, error) {
	var migrationLog MigrationLog
	if err := db.Where("deployment_id = ?", deploymentID).
		Find(&migrationLog).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return &migrationLog, nil
}

// GetMigrationLogExpired 获取已经过期的migration操作记录
func (db *DBClient) GetMigrationLogExpiredThreeDays() (*[]MigrationLog, error) {
	var migrationLogs []MigrationLog
	currentTime := time.Now()
	LastThreeDaysTime := currentTime.AddDate(0, 0, -3)
	if err := db.Where("created_at < ?", LastThreeDaysTime).
		Where("status != ?", apistructs.MigrationStatusDeleted).
		Find(&migrationLogs).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get migration records")
	}

	return &migrationLogs, nil
}
