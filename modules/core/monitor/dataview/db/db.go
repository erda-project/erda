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
	if err := query.Order("created_at DESC").Find(&list).Error; err != nil {
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
	if err := query.Order("created_at DESC").Find(&list).Error; err != nil {
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
