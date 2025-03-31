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
	"errors"

	"github.com/jinzhu/gorm"
)

type MonitorConfigRegisterDB struct {
	*gorm.DB
}

func NewMonitorConfigRegisterDB(db *gorm.DB) *MonitorConfigRegisterDB {
	return &MonitorConfigRegisterDB{db}
}

func (m *MonitorConfigRegisterDB) ListRegisterByOrgId(orgId string) ([]SpMonitorConfigRegister, error) {
	var res = make([]SpMonitorConfigRegister, 0)
	if err := m.Model(&SpMonitorConfigRegister{}).Where("scope = 'org' and scope_id = ?", orgId).Find(&res).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return res, nil
		}
		return nil, err
	}
	return res, nil
}

func (m *MonitorConfigRegisterDB) ListRegisterByType(tpy string) ([]SpMonitorConfigRegister, error) {
	var res = make([]SpMonitorConfigRegister, 0)
	if err := m.Model(&SpMonitorConfigRegister{}).Where("type = ?", tpy).Find(&res).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return res, nil
		}
		return nil, err
	}
	return res, nil
}
