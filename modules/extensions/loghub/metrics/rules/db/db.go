package db

import (
	"github.com/jinzhu/gorm"
)

// DB .
type DB struct {
	*gorm.DB
	LogMetricConfig LogMetricConfigDB
}

// New .
func New(db *gorm.DB) *DB {
	return &DB{
		DB:              db,
		LogMetricConfig: LogMetricConfigDB{db},
	}
}

// Begin .
func (db *DB) Begin() *DB {
	return New(db.DB.Begin())
}
