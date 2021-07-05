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

package db

import (
	"time"

	"github.com/jinzhu/gorm"
)

type MonitorDb struct {
	*gorm.DB
}

func (db *MonitorDb) GetByTerminusKey(terminusKey string) (*Monitor, error) {
	var monitor Monitor
	result := db.Table(TableMonitor).
		Where("`terminus_key`=?", terminusKey).
		Limit(1).
		Find(&monitor)

	if result.RecordNotFound() {
		return nil, nil
	}

	return &monitor, result.Error
}

func (db *MonitorDb) UpdateStatusByMonitorId(monitorId string, delete int) error {
	return db.Table(TableMonitor).
		Where("`monitor_id`=?", monitorId).
		Update(map[string]interface{}{
			"is_delete": delete,
			"updated":   time.Now(),
		}).Error
}
