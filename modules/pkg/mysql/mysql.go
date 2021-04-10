// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	// "github.com/mattn/go-sqlite3"
)

// Query .
type Query interface {
	Supplements(db *gorm.DB) *gorm.DB
}

// GenerateListDb .
func GenerateListDb(
	db *gorm.DB,
	q Query,
	pageSize int64,
	pageNo int64,
) *gorm.DB {
	if pageSize == 0 {
		return q.Supplements(db)
	} else {
		return q.Supplements(db).Limit(pageSize).Offset((pageNo - 1) * pageSize)
	}
}

// GenerateGetDb .
func GenerateGetDb(db *gorm.DB, q Query) *gorm.DB {
	return q.Supplements(db)
}

// IsUniqueConstraintError .
func IsUniqueConstraintError(err error) bool {
	const (
		ErrMySQLDupEntry            = 1062
		ErrMySQLDupEntryWithKeyName = 1586
	)
	/*
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique ||
				sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return true
			}
		} else */
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == ErrMySQLDupEntry ||
			mysqlErr.Number == ErrMySQLDupEntryWithKeyName {
			return true
		}
	}
	return false
}
