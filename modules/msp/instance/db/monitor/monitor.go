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

package monitor

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/database/gormutil"
)

// MoniroeDB .
type MonitorDB struct {
	*gorm.DB
}

func (db *MonitorDB) GetByFields(fields map[string]interface{}) (*Monitor, error) {
	query := db.Table(TableMonitor)
	query, err := gormutil.GetQueryFilterByFields(query, monitorFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*Monitor
	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	return list[0], nil
}

// GetByTerminusKey .
func (db *MonitorDB) GetByTerminusKey(terminusKey string) (*Monitor, error) {
	var monitor Monitor
	result := db.Table(TableMonitor).
		Where("`terminus_key`=?", terminusKey).
		Limit(1).
		Last(&monitor)
	if result.RecordNotFound() {
		return nil, nil
	}
	return &monitor, result.Error
}
