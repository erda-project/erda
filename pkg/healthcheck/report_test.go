package healthcheck

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/jinzhu/gorm"
	"github.com/xormplus/xorm"
	"gotest.tools/assert"

	"terminus.io/dice/dice/pkg/jsonstore"
)

func Test_Report(t *testing.T) {

	var table = []struct {
		dbClient  bool // 是否上报db
		db        bool // 是否上报db
		jsonStore bool // jsonStore
		redisCli  bool // redis

		dbClientResult  bool // 上报的结果
		dbResult        bool // 上报的结果
		jsonStoreResult bool // 上报的结果
		redisCliResult  bool // 上报的结果

		messageStatus Status // 总的结果
	}{
		{
			dbClient:  true,
			db:        true,
			jsonStore: false,

			dbClientResult:  true,
			dbResult:        false,
			jsonStoreResult: true,

			messageStatus: Fail,
		},
		{
			dbClient:  true,
			db:        false,
			jsonStore: false,

			dbClientResult:  true,
			dbResult:        false,
			jsonStoreResult: false,

			messageStatus: Ok,
		},
		{
			dbClient:  false,
			db:        false,
			jsonStore: false,

			dbClientResult:  true,
			dbResult:        true,
			jsonStoreResult: true,

			messageStatus: Ok,
		},
		{
			dbClient:  true,
			db:        true,
			jsonStore: true,

			dbClientResult:  false,
			dbResult:        true,
			jsonStoreResult: true,

			messageStatus: Fail,
		},
		{
			dbClient:  true,
			db:        true,
			jsonStore: true,

			dbClientResult:  true,
			dbResult:        true,
			jsonStoreResult: true,

			messageStatus: Ok,
		},
	}

	for index, data := range table {

		var Engine xorm.Engine
		if data.dbClient {
			mock := monkey.PatchInstanceMethod(reflect.TypeOf(&Engine), "Exec", func(engine *xorm.Engine, sql string, args ...interface{}) (sql.Result, error) {
				if data.dbClientResult {
					return nil, nil
				} else {
					return nil, fmt.Errorf("error")
				}
			})
			RegisterMonitor(MysqlMonitor{
				DbClient: &Engine,
			})
			defer mock.Unpatch()
		}

		var db gorm.DB
		if data.db {
			mock := monkey.PatchInstanceMethod(reflect.TypeOf(&db), "Exec", func(db *gorm.DB, sql string, values ...interface{}) *gorm.DB {
				if data.dbResult {
					return db
				} else {
					db.Error = fmt.Errorf("error")
					return db
				}
			})
			RegisterMonitor(MysqlMonitor{
				Db: &db,
			})
			defer mock.Unpatch()
		}

		var jsonStore, _ = jsonstore.New()
		if data.jsonStore {
			mock := monkey.PatchInstanceMethod(reflect.TypeOf(jsonStore), "Put", func(json *jsonstore.JSONStoreWithWatchImpl, ctx context.Context, key string, object interface{}) error {
				if data.jsonStoreResult {
					return nil
				} else {
					return fmt.Errorf("error")
				}
			})
			RegisterMonitor(JsonStoreMonitor{
				Store: jsonStore,
			})
			defer mock.Unpatch()
		}

		message := DoReport()
		assert.Equal(t, message.Status, data.messageStatus, fmt.Sprintf("index %d: status error", index))
		report.monitors = nil
	}

}
