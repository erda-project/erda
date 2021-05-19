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

package notify_record

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/monitor/notify/template/db"
	"github.com/erda-project/erda/modules/monitor/notify/template/model"
	"github.com/erda-project/erda/modules/monitor/notify/template/query"
)

func (p *provider) invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	notifyRecord := model.NotifyRecord{}
	if err := json.Unmarshal(value, notifyRecord); err != nil {
		return err
	}
	record := query.ToNotifyRecord(&notifyRecord)
	var sqlRecord *db.NotifyRecord
	err := p.mysql.Model(&db.NotifyRecord{}).Where("notify_id = ?", notifyRecord.NotifyId).First(sqlRecord).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if sqlRecord == nil {
		err := p.mysql.Create(record).Error
		return err
	} else {
		err := p.mysql.Model(&db.NotifyRecord{}).Where("notify_id = ?", notifyRecord.NotifyId).Updates(map[string]interface{}{
			"notify_id":   record.NotifyId,
			"notify_name": record.NotifyName,
			"notify_time": record.NotifyTime,
			"title":       record.Title,
		}).Error
		return err
	}
}
