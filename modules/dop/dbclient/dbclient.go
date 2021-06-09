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

// Package dbclient defines operations about database
package dbclient

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"

	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

const BULK_INSERT_CHUNK_SIZE = 3000

var DB *DBClient

type DBClient struct {
	*dbengine.DBEngine
}

func Open() error {
	if DB != nil {
		return nil
	}

	engine, err := dbengine.Open()
	if err != nil {
		return err
	}

	if conf.Debug() {
		engine.LogMode(true)
	}

	DB = &DBClient{DBEngine: engine}

	// custom init
	if err := DB.initOpts(); err != nil {
		return err
	}

	return nil
}

func Close() error {
	if DB == nil || DB.DBEngine == nil {
		return nil
	}
	return DB.DBEngine.Close()
}

func (db *DBClient) initOpts() error {
	return nil
}

func (db *DBClient) BulkInsert(objects interface{}, excludeColumns ...string) error {
	v := reflect.ValueOf(objects)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("invalid objects type, must be a slice of struct")
	}
	var structSlice []interface{}
	for i := 0; i < v.Len(); i++ {
		structSlice = append(structSlice, v.Index(i).Interface())
	}
	return gormbulk.BulkInsert(db.DB, structSlice, BULK_INSERT_CHUNK_SIZE, excludeColumns...)
}

// Transaction Execute Transaction
func (db *DBClient) Transaction(f func(tx *gorm.DB) error) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
