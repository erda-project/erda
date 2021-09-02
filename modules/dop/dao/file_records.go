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
	"database/sql/driver"
	"encoding/json"
	"fmt"
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

type stateCounter struct {
	Type  string
	Count int
}

// Get Records by projectId
func (client *DBClient) ListRecordsByProject(req apistructs.ListTestFileRecordsRequest) ([]TestFileRecord, map[string]int, error) {
	var res []TestFileRecord
	if len(req.Types) > 0 {
		if err := client.Where("`project_id` = ? AND `type` IN (?)", req.ProjectID, req.Types).Order("created_at desc").Find(&res).Error; err != nil {
			return nil, nil, err
		}
	} else {
		if err := client.Where("`project_id` = ?", req.ProjectID).Order("created_at desc").Find(&res).Error; err != nil {
			return nil, nil, err
		}
	}

	var counterList []stateCounter
	if err := client.Table("dice_test_file_records").Select("type, count(*) as count").Where("`type` IN (?) AND `state` = ?", req.Types, apistructs.FileRecordStatePending).Group("type").Find(&counterList).Error; err != nil {
		return nil, nil, err
	}

	counter := make(map[string]int)
	for _, s := range counterList {
		counter[s.Type] = s.Count
	}

	return res, counter, nil
}

// Update Record
func (client *DBClient) UpdateRecord(record *TestFileRecord) error {
	return client.Save(record).Error
}

func (client *DBClient) FirstFileReady(actionType ...apistructs.FileActionType) (bool, *TestFileRecord, error) {
	if len(actionType) == 0 {
		return false, nil, fmt.Errorf("failed to get first file ready, err: empty action type")
	}
	var process int64

	if err := client.Model(&TestFileRecord{}).Where("`state` = ? AND `type` in (?)", apistructs.FileRecordStateProcessing, actionType).Count(&process).Error; err != nil {
		return false, nil, err
	}
	if process > 0 {
		return false, nil, nil
	}

	var record TestFileRecord
	if err := client.Model(&TestFileRecord{}).Where("`state` = ? AND `type` in (?)", apistructs.FileRecordStatePending, actionType).Order("created_at").First(&record).Error; err != nil {
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
