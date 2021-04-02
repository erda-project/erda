package dao

import (
	"fmt"
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"

	"github.com/erda-project/erda/modules/cmdb/conf"
)

// DIALECT db 类型
const DIALECT = "mysql"

const BULK_INSERT_CHUNK_SIZE = 3000

// DBClient db client
type DBClient struct {
	*gorm.DB
}

func newDB() (*gorm.DB, error) {
	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		conf.MySQLUsername(), conf.MySQLPassword(), conf.MySQLHost(), conf.MySQLPort(), conf.MySQLDatabase(), conf.MySQLLoc())

	logrus.Debugf("Initialize db with %s, url: %s", DIALECT, url)

	db, err := gorm.Open(DIALECT, url)
	if err != nil {
		return nil, err
	}
	if conf.Debug() {
		db.LogMode(true)
	}
	// connection pool
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(50)
	db.DB().SetConnMaxLifetime(time.Hour)

	return db, nil
}

// NewDBClient create new db client
func NewDBClient() (*DBClient, error) {
	var err error

	client := &DBClient{}
	client.DB, err = newDB()

	return client, err
}

func (client *DBClient) BulkInsert(objects interface{}, excludeColumns ...string) error {
	v := reflect.ValueOf(objects)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("invalid objects type, must be a slice of struct")
	}
	var structSlice []interface{}
	for i := 0; i < v.Len(); i++ {
		structSlice = append(structSlice, v.Index(i).Interface())
	}
	return gormbulk.BulkInsert(client.DB, structSlice, BULK_INSERT_CHUNK_SIZE, excludeColumns...)
}
