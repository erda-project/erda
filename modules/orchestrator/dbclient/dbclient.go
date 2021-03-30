// Package dbclient 定义数据库操作的方法, orm 等。
package dbclient

import (
	"github.com/erda-project/erda/pkg/dbengine"
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
