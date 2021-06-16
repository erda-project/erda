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
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/database/gormutil"
)

// InstanceDB .
type InstanceDB struct {
	*gorm.DB
}

func (db *InstanceDB) GetByID(id string) (*Instance, error) {
	if len(id) < 0 {
		return nil, nil
	}
	var list []*Instance
	if err := db.Table(TableInstance).
		Where("`id`=?", id).Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	return list[0], nil
}

func (db *InstanceDB) GetByFields(fields map[string]interface{}) (*Instance, error) {
	query := db.Table(TableInstance)
	query, err := gormutil.GetQueryFilterByFields(query, instanceFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*Instance
	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	return list[0], nil
}
