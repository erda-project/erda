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
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/core-services/model"
)

// CreateSubscribe Create relationship between erda item (project/application) and subscriber
func (client *DBClient) CreateSubscribe(subscribe *model.Subscribe) error {
	return client.Create(subscribe).Error
}

// DeleteSubscribe Delete subscribe relation
func (client *DBClient) DeleteSubscribe(tp string, tpID uint64, userID string) error {
	return client.Where("type = ? and type_id = ? and user_id = ?", tp, tpID, userID).Delete(&model.Subscribe{}).Error
}

// DeleteSubscribeByTypeID Delete subscribe by type id
func (client *DBClient) DeleteSubscribeByTypeID(tp string, tpID uint64) error {
	return client.Where("type = ? and type_id = ?", tp, tpID).Delete(&model.Subscribe{}).Error
}

// DeleteSubscribeByUserID Delete subscribe by user id
func (client *DBClient) DeleteSubscribeByUserID(userID string) error {
	return client.Where("user_id = ?", userID).Delete(&model.Subscribe{}).Error
}

func (client *DBClient) GetSubscribeCount(tp string, userID string) (int, error) {
	var total int
	err := client.Model(&model.Subscribe{}).Where("type = ? and user_id = ?", tp, userID).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetSubscribe get subscribe
func (client *DBClient) GetSubscribe(tp string, tpID uint64, userID string) (*model.Subscribe, error) {
	var subscribe model.Subscribe
	if err := client.Model(model.Subscribe{}).Where("type = ? and type_id = ? and user_id = ?", tp,
		tpID, userID).First(&subscribe).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}

		return nil, err
	}

	return &subscribe, nil
}

// GetSubscribesByUserID get subscribes by user id
func (client *DBClient) GetSubscribesByUserID(tp string, userID string) ([]model.Subscribe, error) {
	var subscribes []model.Subscribe
	if err := client.Model(model.Subscribe{}).Where("type = ? and user_id = ?", tp,
		userID).Find(&subscribes).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}

		return nil, err
	}

	return subscribes, nil
}
