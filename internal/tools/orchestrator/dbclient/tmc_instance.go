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

package dbclient

import (
	"time"

	"github.com/jinzhu/gorm"
)

const (
	TMCInstanceRunningStatus = "RUNNING"
	TMCInstanceErrorStatus   = "ERROR"
	TMCInstanceInitStatus    = "INIT"
)

// TmcInstance .
type TmcInstance struct {
	ID         string    `gorm:"column:id;primary_key"`
	Engine     string    `gorm:"column:engine"`
	Version    string    `gorm:"column:version"`
	ReleaseID  string    `gorm:"column:release_id"`
	Status     string    `gorm:"column:status"`
	Az         string    `gorm:"column:az"`
	Config     string    `gorm:"column:config"`
	Options    string    `gorm:"column:options"`
	IsCustom   string    `gorm:"column:is_custom;default:'N'"`
	IsDeleted  string    `gorm:"column:is_deleted;default:'N'"`
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}

// TableName .
func (TmcInstance) TableName() string { return "tb_tmc_instance" }

func (db *DBClient) FindTmcInstanceByNameAndCLuster(name, cluster string) ([]TmcInstance, error) {
	var res []TmcInstance
	model := db.Model(&TmcInstance{})
	if err := model.
		Where("engine = ?", name).
		Where("az = ?", cluster).
		Where("status = ?", TMCInstanceRunningStatus).
		Where("is_deleted = 'N'").
		Find(&res).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return res, nil
		}
		return nil, err
	}
	return res, nil
}
