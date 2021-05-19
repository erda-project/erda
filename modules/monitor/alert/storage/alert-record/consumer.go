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

package alert_record

import (
	"encoding/json"
	"time"

	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"

	"github.com/jinzhu/gorm"
)

func (p *provider) invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	p.mysql.LogMode(true)
	alertRecord := &adapt.AlertRecord{}
	if err := json.Unmarshal(value, alertRecord); err != nil {
		return err
	}
	record := &db.AlertRecord{}
	alertRecord.ToModel(record)
	var sqlRecord *db.AlertRecord
	err := p.mysql.Model(&db.AlertRecord{}).Where("group_id = ?", alertRecord.GroupID).First(sqlRecord).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if sqlRecord == nil {
		//create
		err := p.mysql.Create(record).Error
		return err
	} else {
		//update
		err := p.mysql.Model(&db.AlertRecord{}).Where("group_id = ?", alertRecord.GroupID).Updates(map[string]interface{}{
			"alert_state": record.AlertState,
			"alert_type":  record.AlertType,
			"alert_index": record.AlertIndex,
			"alert_name":  record.AlertName,
			"alert_time":  record.AlertTime,
		}).Error
		return err
	}
}
