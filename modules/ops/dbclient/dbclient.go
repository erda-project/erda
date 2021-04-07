// Package dbclient 定义数据库操作的方法, orm 等。
package dbclient

import (
	"github.com/erda-project/erda/pkg/dbengine"
)

type DBClient struct {
	*dbengine.DBEngine
}

func Open(db *dbengine.DBEngine) *DBClient {
	return &DBClient{DBEngine: db}
}
