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

package dbengine

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DBEngine struct {
	*gorm.DB
}

// Open 构造引擎，打开数据库连接池
func Open(c ...*Conf) (*DBEngine, error) {
	cfg, err := loadCfg(c...)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open("mysql", cfg.url())
	if err != nil {
		return nil, err
	}
	// connection pool
	db.DB().SetMaxIdleConns(cfg.maxIdleConns())
	db.DB().SetMaxOpenConns(cfg.maxOpenConns())
	db.DB().SetConnMaxLifetime(cfg.maxLifeTime())

	// debug
	if cfg.Debug {
		db.LogMode(true)
	}

	engine := DBEngine{
		DB: db,
	}
	return &engine, nil
}

// MustOpen 强制打开，err 时 panic
func MustOpen(c ...*Conf) *DBEngine {
	engine, err := Open(c...)
	if err != nil {
		panic(err)
	}
	return engine
}

// Close 关闭数据库连接池
func (e *DBEngine) Close() error {
	if e == nil || e.DB == nil {
		return nil
	}
	return e.DB.Close()
}

func loadCfg(c ...*Conf) (*Conf, error) {
	if len(c) > 0 {
		return c[0], nil
	}
	return LoadDefaultConf()
}
