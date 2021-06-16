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

package gormutil

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

// GetFieldToColumnMap .
func GetFieldToColumnMap(typ reflect.Type) map[string]string {
	if typ.Kind() != reflect.Struct {
		return nil
	}
	m := make(map[string]string)
loop:
	for i, n := 0, typ.NumField(); i < n; i++ {
		field := typ.Field(i)
		dbtag := field.Tag.Get("gorm")
		if len(dbtag) > 0 {
			for _, item := range strings.Split(dbtag, ";") {
				kv := strings.SplitN(item, ":", 2)
				if len(kv) == 2 {
					k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
					if k == "column" && len(v) > 0 {
						m[field.Name] = v
						continue loop
					}
				}
			}
		}
		m[field.Name] = strings.ToLower(field.Name)
	}
	return m
}

// GetQueryFilterByFields .
func GetQueryFilterByFields(db *gorm.DB, fieldColumns map[string]string, fields map[string]interface{}) (*gorm.DB, error) {
	for name, value := range fields {
		col, ok := fieldColumns[name]
		if !ok {
			return nil, fmt.Errorf("unknown %q", name)
		}
		db = db.Where(fmt.Sprintf("`%s`=?", col), value)
	}
	return db, nil
}

// GetQueryFilterByFields .
func GetQueryFilterByTypeFields(db *gorm.DB, typ interface{}, fields map[string]interface{}) (*gorm.DB, error) {
	type TableNamer interface {
		TableName() string
	}
	if table, ok := typ.(TableNamer); ok {
		db = db.Table(table.TableName())
	}
	return GetQueryFilterByFields(db, GetFieldToColumnMap(reflect.TypeOf(typ)), fields)
}
