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
	"github.com/jinzhu/gorm"

	db2 "github.com/erda-project/erda/modules/tools/monitor/core/alert/alert-apis/db"
)

type DB struct {
	*gorm.DB
	AlertNotifyIndexDB AlertNotifyIndexDB
	NotifyHistoryDB    NotifyHistoryDB
	AlertNotifyDB      db2.AlertNotifyDB
}

func New(db *gorm.DB) *DB {
	return &DB{
		DB:                 db,
		AlertNotifyIndexDB: AlertNotifyIndexDB{db},
		NotifyHistoryDB:    NotifyHistoryDB{db},
		AlertNotifyDB:      db2.AlertNotifyDB{db},
	}
}

func (db *DB) Begin() *DB {
	return New(db.DB.Begin())
}
