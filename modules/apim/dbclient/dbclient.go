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
