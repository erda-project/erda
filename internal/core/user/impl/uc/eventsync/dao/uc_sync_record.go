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
	"fmt"
	"reflect"
	"time"

	gormbulk "github.com/t-tiger/gorm-bulk-insert"

	"github.com/erda-project/erda/pkg/database/dbengine"
)

// UCSyncRecord holds UC event sync history.
type UCSyncRecord struct {
	dbengine.BaseModel
	UCID        int64     `gorm:"column:uc_id"`
	UCEventTime time.Time `gorm:"column:uc_eventtime"`
	UnReceiver  string    `gorm:"un_receiver"` // receivers that failed to sync, for compensation
}

func (UCSyncRecord) TableName() string {
	return "dice_ucevent_sync_record"
}

// BulkInsert inserts objects in batch.
func (client *DBClient) BulkInsert(objects interface{}, excludeColumns ...string) error {
	v := reflect.ValueOf(objects)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("invalid objects type, must be a slice of struct")
	}
	var structSlice []interface{}
	for i := 0; i < v.Len(); i++ {
		structSlice = append(structSlice, v.Index(i).Interface())
	}
	return gormbulk.BulkInsert(client.DB, structSlice, bulkInsertChunkSize, excludeColumns...)
}

// BatchCreateUCSyncRecord batch creates UC audit sync records.
func (client *DBClient) BatchCreateUCSyncRecord(records []*UCSyncRecord) error {
	if len(records) == 0 {
		return nil
	}
	return client.BulkInsert(records)
}

// CreateUCSyncRecord creates a sync record.
func (client *DBClient) CreateUCSyncRecord(ucSyncRecord *UCSyncRecord) error {
	return client.Create(ucSyncRecord).Error
}

// UpdateRecord updates the record.
func (client *DBClient) UpdateRecord(record *UCSyncRecord) error {
	return client.Table("dice_ucevent_sync_record").Save(record).Error
}

// DeleteRecordByTime deletes sync records before the given time.
func (client *DBClient) DeleteRecordByTime(t time.Time) error {
	return client.Table("dice_ucevent_sync_record").Where("uc_eventtime < ?", t).Delete(&UCSyncRecord{}).Error
}

// GetLastNRecord returns the last N sync records.
func (client *DBClient) GetLastNRecord(n int) ([]UCSyncRecord, error) {
	var ucSyncRecords []UCSyncRecord
	if err := client.Table("dice_ucevent_sync_record").Order("uc_id DESC").Offset(0).Limit(n).Find(&ucSyncRecords).Error; err != nil {
		return nil, err
	}
	return ucSyncRecords, nil
}

// GetRecordByUCIDs returns sync records by UC IDs.
func (client *DBClient) GetRecordByUCIDs(ucid []int64) ([]UCSyncRecord, error) {
	var ucSyncRecords []UCSyncRecord
	if err := client.Table("dice_ucevent_sync_record").Where("uc_id in (?)", ucid).Find(&ucSyncRecords).Error; err != nil {
		return nil, err
	}
	return ucSyncRecords, nil
}

// GetFailedRecord returns records that failed to send.
func (client *DBClient) GetFailedRecord(size int) ([]UCSyncRecord, error) {
	var ucSyncRecords []UCSyncRecord
	if err := client.Table("dice_ucevent_sync_record").Where("un_receiver != ?", "").Order("uc_id ASC").Offset(0).
		Limit(size).Find(&ucSyncRecords).Error; err != nil {
		return nil, err
	}
	return ucSyncRecords, nil
}
