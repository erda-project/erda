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

// Package dbclient 定义数据库操作的方法, orm 等。
package dbclient

import (
	"github.com/erda-project/erda/pkg/dbengine"
)

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

// TODO: 自定义初始化内容
func (db *DBClient) initOpts() error {
	return nil
}
