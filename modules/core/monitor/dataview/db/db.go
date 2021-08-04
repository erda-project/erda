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
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/database/gormutil"
)

// SystemViewDB .
type SystemViewDB struct {
	*gorm.DB
}

func (db *SystemViewDB) query() *gorm.DB      { return db.Table(TableSystemView) }
func (db *SystemViewDB) Begin() *SystemViewDB { return &SystemViewDB{DB: db.DB.Begin()} }

func (db *SystemViewDB) GetByFields(fields map[string]interface{}) (*SystemView, error) {
	query, err := gormutil.GetQueryFilterByFields(db.query(), systemViewFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*SystemView
	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	return list[0], nil
}

func (db *SystemViewDB) ListByFields(fields map[string]interface{}) ([]*SystemView, error) {
	query, err := gormutil.GetQueryFilterByFields(db.query(), systemViewFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*SystemView
	if err := query.Find(&list).Order("created_at DESC").Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CustomViewDB .
type CustomViewDB struct {
	*gorm.DB
}

func (db *CustomViewDB) query() *gorm.DB      { return db.Table(TableCustomView) }
func (db *CustomViewDB) Begin() *CustomViewDB { return &CustomViewDB{DB: db.DB.Begin()} }

func (db *CustomViewDB) GetByFields(fields map[string]interface{}) (*CustomView, error) {
	query, err := gormutil.GetQueryFilterByFields(db.query(), customViewFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*CustomView
	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	return list[0], nil
}

func (db *CustomViewDB) ListByFields(fields map[string]interface{}) ([]*CustomView, error) {
	query, err := gormutil.GetQueryFilterByFields(db.query(), customViewFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*CustomView
	if err := query.Find(&list).Order("created_at DESC").Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (db *CustomViewDB) UpdateView(id string, fields map[string]interface{}) error {
	updates := make(map[string]interface{})
	for name, value := range fields {
		col, ok := customViewFieldColumns[name]
		if !ok {
			return fmt.Errorf("unknown %q", name)
		}
		updates[col] = value
	}
	return db.query().Where("id=?", id).Updates(updates).Error
}
