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

package dbclient

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/apim/conf"
)

type TX struct {
	*gorm.DB
}

func (tx *TX) Sq() *gorm.DB {
	if tx == nil {
		return nil
	}
	return tx.DB
}

func (db *DBClient) Tx() *TX {
	if conf.Debug() {
		return &TX{DB: db.Begin().Debug()}
	}
	return &TX{DB: db.Begin()}
}

func (db *DBClient) Sq() *gorm.DB {
	if conf.Debug() {
		return db.Debug()
	}
	return db.DB
}

func Tx() *TX {
	return DB.Tx()
}

func Sq() *gorm.DB {
	return DB.Sq()
}
