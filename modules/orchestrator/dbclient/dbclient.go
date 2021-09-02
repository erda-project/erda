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

// Package dbclient 定义数据库操作的方法, orm 等。
package dbclient

import (
	"github.com/erda-project/erda/pkg/database/dbengine"
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
