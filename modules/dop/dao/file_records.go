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
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type TestFileRecord struct {
	dbengine.BaseModel
	FileName    string
	Description string
	ApiFileUUID string
	ProjectID   uint64
	Type        apistructs.FileActionType
	State       apistructs.FileRecordState
	OperatorID  string
	Extra       TestFileExtra
}

type TestFileExtra apistructs.TestFileExtra

func (ex TestFileExtra) Value() (driver.Value, error) {
	if b, err := json.Marshal(ex); err != nil {
		return nil, errors.Wrapf(err, "failed to marshal Extra")
	} else {
		return string(b), nil
	}
}

func (ex *TestFileExtra) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for Extra")
	}
	if len(v) == 0 {
		return nil
	}
	if err := json.Unmarshal(v, ex); err != nil {
		return errors.Wrapf(err, "failed to unmarshal Extra")
	}
	return nil
}

// Test TableName
func (TestFileRecord) TableName() string {
	return "dice_test_file_records"
}

// Create Record
func (client *DBClient) CreateRecord(record *TestFileRecord) error {
	return client.Create(record).Error
}

// Get Record by id
func (client *DBClient) GetRecord(id uint64) (*TestFileRecord, error) {
	var res TestFileRecord
	if err := client.First(&res, id).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

// Get Records by projectId
func (client *DBClient) ListRecordsByProject(req apistructs.ListTestFileRecordsRequest) ([]TestFileRecord, error) {
	var res []TestFileRecord
	if len(req.Types) > 0 {
		if err := client.Where("`project_id` = ? AND `type` IN (?)", req.ProjectID, req.Types).Order("created_at desc").Find(&res).Error; err != nil {
			return nil, err
		}
	} else {
		if err := client.Where("`project_id` = ?", req.ProjectID).Order("created_at desc").Find(&res).Error; err != nil {
			return nil, err
		}
	}
	return res, nil
}

// Update Record
func (client *DBClient) UpdateRecord(record *TestFileRecord) error {
	return client.Save(record).Error
}

func (client *DBClient) FirstFileReady(actionType apistructs.FileActionType) (bool, *TestFileRecord, error) {
	var process int64

	if err := client.Model(&TestFileRecord{}).Where("`state` = ? AND `type` = ?", apistructs.FileRecordStateProcessing, actionType).Count(&process).Error; err != nil {
		return false, nil, err
	}
	if process > 0 {
		return false, nil, nil
	}

	var record TestFileRecord
	if err := client.Model(&TestFileRecord{}).Where("`state` = ? AND `type` = ?", apistructs.FileRecordStatePending, actionType).Order("created_at").First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}

	return true, &record, nil
}

func (client *DBClient) BatchUpdateRecords() error {
	return client.Model(&TestFileRecord{}).Where("`state` = ?", apistructs.FileRecordStateProcessing).Updates(TestFileRecord{State: apistructs.FileRecordStateFail}).Error
}

type FileUUIDStr struct {
	ApiFileUUID string
}

func (client *DBClient) DeleteFileRecordByTime(t time.Time) ([]FileUUIDStr, error) {
	var res []FileUUIDStr
	if err := client.Table("dice_test_file_records").Where("`created_at` < ?", t).Select("api_file_uuid").Find(&res).Error; err != nil {
		return nil, err
	}

	if err := client.Where("`created_at` < ?", t).Delete(TestFileRecord{}).Error; err != nil {
		return nil, err
	}

	return res, nil
}
