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
	"reflect"
	"time"

	gormbulk "github.com/t-tiger/gorm-bulk-insert"

	"github.com/erda-project/erda/pkg/dbengine"
)

// UCSyncRecord uc同步历史记录
type UCSyncRecord struct {
	dbengine.BaseModel
	UCID        int64     `gorm:"column:uc_id"`
	UCEventTime time.Time `gorm:"column:uc_eventtime"`
	UnReceiver  string    `gorm:"un_receiver"` // uc事件同步失败的补偿标记
}

func (UCSyncRecord) TableName() string {
	return "dice_ucevent_sync_record"
}

// BulkInsert 批量插入
func (client *DBClient) BulkInsert(objects interface{}, excludeColumns ...string) error {
	v := reflect.ValueOf(objects)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("invalid objects type, must be a slice of struct")
	}
	var structSlice []interface{}
	for i := 0; i < v.Len(); i++ {
		structSlice = append(structSlice, v.Index(i).Interface())
	}
	return gormbulk.BulkInsert(client.DB, structSlice, BULK_INSERT_CHUNK_SIZE, excludeColumns...)
}

// BatchCreateUCSyncRecord 批量插入 uc 审计同步记录
func (client *DBClient) BatchCreateUCSyncRecord(records []*UCSyncRecord) error {
	return client.BulkInsert(records)
}

// CreateUCSyncRecord 创建同步历史记录
func (client *DBClient) CreateUCSyncRecord(ucSyncRecord *UCSyncRecord) error {
	return client.Create(ucSyncRecord).Error
}

// UpdateRecord 更新历史记录
func (client *DBClient) UpdateRecord(record *UCSyncRecord) error {
	return client.Table("dice_ucevent_sync_record").Save(record).Error
}

// DeleteRecordByTime 根据时间删除同步记录
func (client *DBClient) DeleteRecordByTime(t time.Time) error {
	return client.Table("dice_ucevent_sync_record").Where("uc_eventtime < ?", t).Delete(&UCSyncRecord{}).Error
}

// GetLastNRecord 获取最后N条同步记录
func (client *DBClient) GetLastNRecord(n int) ([]UCSyncRecord, error) {
	var ucSyncRecords []UCSyncRecord
	if err := client.Table("dice_ucevent_sync_record").Order("uc_id DESC").Offset(0).Limit(n).Find(&ucSyncRecords).Error; err != nil {
		return nil, err
	}

	return ucSyncRecords, nil
}

// GetRecordByUCIDs 根据ucid获取同步记录
func (client *DBClient) GetRecordByUCIDs(ucid []int64) ([]UCSyncRecord, error) {
	var ucSyncRecords []UCSyncRecord
	if err := client.Table("dice_ucevent_sync_record").Where("uc_id in (?)", ucid).Find(&ucSyncRecords).Error; err != nil {
		return nil, err
	}

	return ucSyncRecords, nil
}

// GetFaieldRecord 获取发送失败的记录
func (client *DBClient) GetFaieldRecord(size int) ([]UCSyncRecord, error) {
	var ucSyncRecords []UCSyncRecord
	if err := client.Table("dice_ucevent_sync_record").Where("un_receiver != ?", "").Order("uc_id ASC").Offset(0).
		Limit(size).Find(&ucSyncRecords).Error; err != nil {
		return nil, err
	}

	return ucSyncRecords, nil
}
