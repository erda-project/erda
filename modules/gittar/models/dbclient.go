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

package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/database/dbengine"
)

type DBClient struct {
	*dbengine.DBEngine
}

func OpenDB() (*DBClient, error) {
	engine, err := dbengine.Open()
	if err != nil {
		return nil, err
	}
	db := DBClient{DBEngine: engine}
	// custom init
	if err := db.initOpts(); err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *DBClient) Close() error {
	if db == nil || db.DBEngine == nil {
		return nil
	}
	return db.DBEngine.Close()
}

func (db *DBClient) initOpts() error {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "dice_repo_" + defaultTableName
	}
	err := db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").Error
	if err != nil {
		logrus.Errorf("db migrate error %v", err)
		return err
	}
	return nil
}
