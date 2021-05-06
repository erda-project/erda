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

type DBClient struct {
	*dbengine.DBEngine
}

func Open() (*DBClient, error) {
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

// TODO: 重构
func (db *DBClient) initOpts() error {
	// set tables prefix: ps_
	// gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
	// 	if strings.HasPrefix(defaultTableName, "ps_") {
	// 		return defaultTableName
	// 	}
	// 	return "ps_" + defaultTableName
	// }
	// db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").
	// 	AutoMigrate(
	// 		&Runtime{},
	// 		&RuntimeService{},
	// 		&RuntimeInstance{},
	// 		&RuntimeDomain{},
	// 		&Deployment{},
	// 		&PreDeployment{},
	// 		&AddonInstance{},
	// 		&AddonInstanceRouting{},
	// 		&AddonAttachment{},
	// 		&AddonAudit{},
	// 		&AddonDeploy{},
	// 		&AddonExtra{},
	// 		&AddonMicroAttach{},
	// 		&AddonPrebuild{},
	// 	)
	return nil
}
