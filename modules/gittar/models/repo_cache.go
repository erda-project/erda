package models

import "time"

type RepoCache struct {
	ID        int64
	TypeName  string `gorm:"size:150;index:type_name"`
	KeyName   string `gorm:"size:150;index:key_name"`
	Value     string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (RepoCache) TableName() string {
	return "dice_repo_caches"
}
