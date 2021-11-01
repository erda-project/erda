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

package db

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/common/errors"
)

// NotifyChannelDB erda_notify_channel
type NotifyChannelDB struct {
	*gorm.DB
}

func (db *NotifyChannelDB) db() *gorm.DB {
	return db.Table(model.TableNotifyChannel)
}

func (db *NotifyChannelDB) Create(notifyChannel *model.NotifyChannel) (*model.NotifyChannel, error) {
	err := db.db().Create(notifyChannel).Error
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return notifyChannel, nil
}

func (db *NotifyChannelDB) GetCountByScopeAndType(scopeId, scopeType, channelType string) (int64, error) {
	var count int64
	err := db.db().
		Where("`scope_id` = ?", scopeId).
		Where("`scope_type` = ?", scopeType).
		Where("`type` = ?", channelType).
		Where("`is_enabled` = ?", true).
		Where("`is_deleted` = ?", false).
		Count(&count).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, errors.NewDatabaseError(err)
	}
	return count, err
}

func (db *NotifyChannelDB) GetByScopeAndType(scopeId, scopeType, channelType string) (*model.NotifyChannel, error) {
	channel := &model.NotifyChannel{}
	err := db.db().
		Where("`scope_id` = ?", scopeId).
		Where("`scope_type` = ?", scopeType).
		Where("`type` = ?", channelType).
		Where("`is_enabled` = ?", true).
		Where("`is_deleted` = ?", false).
		Find(channel).
		Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return channel, nil
}

func (db *NotifyChannelDB) GetById(id string) (*model.NotifyChannel, error) {
	channel := &model.NotifyChannel{}
	err := db.db().Where("`id` = ?", id).Where("`is_deleted` = ?", false).Find(channel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return channel, nil
}

func (db *NotifyChannelDB) GetByName(name string) (*model.NotifyChannel, error) {
	channel := &model.NotifyChannel{}
	err := db.db().Where("`name` = ?", name).Where("`is_deleted` = ?", false).Find(channel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return channel, nil
}

func (db *NotifyChannelDB) GetCountByName(name string) (int64, error) {
	var count int64
	err := db.db().Where("`name` = ?", name).Where("`is_deleted` = ?", false).Count(&count).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, errors.NewDatabaseError(err)
	}
	return count, nil
}

func (db *NotifyChannelDB) DeleteById(id string) (*model.NotifyChannel, error) {
	channel, err := db.GetById(id)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if channel == nil {
		return nil, errors.NewDatabaseError(errors.NewNotFoundError(id))
	}
	channel.IsDeleted = true
	err = db.db().Model(channel).Update(channel).Error
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return channel, nil
}

func (db *NotifyChannelDB) UpdateById(notifyChannel *model.NotifyChannel) (*model.NotifyChannel, error) {

	err := db.db().Model(notifyChannel).Save(notifyChannel).Error
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return notifyChannel, nil
}

func (db *NotifyChannelDB) ListByPage(offset, pageSize int64, scopeId, scopeType string) (int64, []model.NotifyChannel, error) {
	var channels []model.NotifyChannel
	whereDB := db.db().
		Where("`is_deleted` = ?", false).
		Where("`scope_id` = ?", scopeId).
		Where("`scope_type` = ?", scopeType)

	err := whereDB.
		Order("`updated_at` DESC", true).
		Offset(offset).
		Limit(pageSize).
		Find(&channels).
		Error
	if err != nil {
		return 0, nil, errors.NewDatabaseError(err)
	}
	var totalCount int64 = 0
	err = whereDB.Count(&totalCount).Error
	if err != nil {
		return 0, nil, errors.NewDatabaseError(err)
	}

	return totalCount, channels, nil
}
