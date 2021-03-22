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
