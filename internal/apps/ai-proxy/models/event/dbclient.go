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

package event

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type DBClient struct {
	DB *gorm.DB
}

type ListOptions struct {
	PageNum  uint64
	PageSize uint64
	Day      string
}

func (dbClient *DBClient) Create(ctx context.Context, eventName, detail string) (*Event, error) {
	rec := &Event{Event: eventName, Detail: detail}
	if err := dbClient.DB.WithContext(ctx).Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

func (dbClient *DBClient) LatestByEvent(ctx context.Context, eventName string) (*Event, error) {
	rec := &Event{}
	err := dbClient.DB.WithContext(ctx).
		Where("event = ?", eventName).
		Order("created_at DESC, id DESC").
		First(rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (dbClient *DBClient) LatestByEvents(ctx context.Context, eventNames ...string) (*Event, error) {
	if len(eventNames) == 0 {
		return nil, nil
	}
	rec := &Event{}
	err := dbClient.DB.WithContext(ctx).
		Where("event IN ?", eventNames).
		Order("created_at DESC, id DESC").
		First(rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (dbClient *DBClient) LatestByEventAndDetail(ctx context.Context, eventName, detail string) (*Event, error) {
	rec := &Event{}
	err := dbClient.DB.WithContext(ctx).
		Where("event = ? AND detail = ?", eventName, detail).
		Order("created_at DESC, id DESC").
		First(rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (dbClient *DBClient) ListDayEvents(ctx context.Context, opt ListOptions) (int64, Events, error) {
	if opt.PageNum == 0 {
		opt.PageNum = 1
	}
	if opt.PageSize == 0 {
		opt.PageSize = 10
	}
	if opt.PageSize > 100 {
		opt.PageSize = 100
	}

	sql := dbClient.DB.WithContext(ctx).
		Model(&Event{}).
		Where("event IN ?", []string{
			EventArchiveDayStart,
			EventArchiveDaySuccess,
			EventArchiveDayFailed,
			EventArchiveDayInterrupted,
			EventArchiveDayEnd,
		})
	if opt.Day != "" {
		sql = sql.Where("detail = ?", opt.Day)
	}

	var (
		total int64
		list  Events
	)
	offset := (opt.PageNum - 1) * opt.PageSize
	if err := sql.Count(&total).Order("created_at DESC, id DESC").Offset(int(offset)).Limit(int(opt.PageSize)).Find(&list).Error; err != nil {
		return 0, nil, err
	}
	return total, list, nil
}
