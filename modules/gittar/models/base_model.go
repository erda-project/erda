package models

import "time"

// Model base model definition, including fields `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`, which could be embedded in your models
//    type User struct {
//      gorm.Model
//    }
type BaseModel struct {
	ID        int64      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}
