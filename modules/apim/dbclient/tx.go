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
