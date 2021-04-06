package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/dbengine"
)

type DBClient struct {
	*dbengine.DBEngine
}

func OpenDB() (*DBClient, error) {
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

func (db *DBClient) Close() error {
	if db == nil || db.DBEngine == nil {
		return nil
	}
	return db.DBEngine.Close()
}

func (db *DBClient) initOpts() error {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "dice_repo_" + defaultTableName
	}
	err := db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").Error
	if err != nil {
		logrus.Errorf("db migrate error %v", err)
		return err
	}
	return nil
}
