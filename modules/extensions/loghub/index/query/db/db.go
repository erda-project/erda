package db

import (
	"github.com/jinzhu/gorm"
)

// DB .
type DB struct {
	*gorm.DB
	LogDeployment LogDeploymentDB
}

// New .
func New(db *gorm.DB) *DB {
	return &DB{
		DB:            db,
		LogDeployment: LogDeploymentDB{db},
	}
}

// Begin .
func (db *DB) Begin() *DB {
	return New(db.DB.Begin())
}
