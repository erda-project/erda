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

package audit

import (
	"context"
	"time"
)

func (dbClient *DBClient) OldestDayBefore(ctx context.Context, cutoff time.Time) (time.Time, bool, error) {
	var rec Audit
	err := dbClient.DB.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Order("created_at ASC, id ASC").
		Limit(1).
		Find(&rec).Error
	if err != nil {
		return time.Time{}, false, err
	}
	if rec.ID.String == "" {
		return time.Time{}, false, nil
	}
	return dayStart(rec.CreatedAt), true, nil
}

func (dbClient *DBClient) HasRowsInRange(ctx context.Context, start, end time.Time) (bool, error) {
	var count int64
	err := dbClient.DB.WithContext(ctx).
		Model(&Audit{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (dbClient *DBClient) ListArchiveBatch(ctx context.Context, start, end time.Time, afterCreatedAt *time.Time, afterID string, limit int) (Audits, error) {
	if limit <= 0 {
		limit = 1000
	}
	sql := dbClient.DB.WithContext(ctx).
		Model(&Audit{}).
		Where("created_at >= ? AND created_at < ?", start, end)
	if afterCreatedAt != nil {
		sql = sql.Where("(created_at > ?) OR (created_at = ? AND id > ?)", *afterCreatedAt, *afterCreatedAt, afterID)
	}

	var list Audits
	err := sql.Order("created_at ASC, id ASC").Limit(limit).Find(&list).Error
	return list, err
}

func (dbClient *DBClient) DeleteArchiveBatch(ctx context.Context, start, end time.Time, limit int) (int64, error) {
	if limit <= 0 {
		limit = 1000
	}
	var rows []struct {
		ID string `gorm:"column:id"`
	}
	if err := dbClient.DB.WithContext(ctx).
		Model(&Audit{}).
		Select("id").
		Where("created_at >= ? AND created_at < ?", start, end).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	tx := dbClient.DB.WithContext(ctx).Unscoped().Where("id IN ?", ids).Delete(&Audit{})
	return tx.RowsAffected, tx.Error
}

func dayStart(t time.Time) time.Time {
	tt := t.In(time.Local)
	return time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
}
