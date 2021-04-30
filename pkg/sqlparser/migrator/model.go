// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package migrator

import (
	"time"
)

const SchemaMigrationHistory = `schema_migration_history`

type HistoryModel struct {
	ID        uint64    `gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt" gorm:"created_at;type:datetime;not null;default CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"updated_at;type:datetime;not null;default CURRENT_TIMESTAMP; on update CURRENT_TIMESTAMP"`

	ServiceName  string `json:"serviceName"  gorm:"service_name;type:varchar(32);not null;comment:微服务名"`
	Filename     string `json:"filename"     gorm:"filename;type:varchar(191);not null;comment:脚本类型"`
	Checksum     string `json:"checksum"     gorm:"checksum;type:char(128);not null;comment:rawtext 的 md5 值"`
	InstalledBy  string `json:"installedBy"  gorm:"installed_by;type:varchar(32);not null;comment:执行人"`
	InstalledOn  string `json:"installedOn"  gorm:"install_on;type:varchar(32);not null;comment:执行平台"`
	LanguageType string `json:"languageType" gorm:"language_type;type:varchar(16);not null;comment:SQL, Go, Java"`
	Reversed     string `json:"reversed"     gorm:"reversed;type:longtext;not null;comment:反转的 DDL"`
}

func (m HistoryModel) TableName() string {
	return SchemaMigrationHistory
}

func (m HistoryModel) create() {
	if ok := DB().Migrator().HasTable(m.TableName()); !ok {
		_ = DB().Migrator().CreateTable(&m)
	}
}

func (m *HistoryModel) insert() error {
	return DB().Create(m).Error
}

func (m *HistoryModel) delete(query interface{}, args ...interface{}) error {
	return DB().Where(query, args...).Delete(m).Error
}

func (m HistoryModel) exists() bool {
	return DB().Raw("SHOW TABLES LIKE ?", m.TableName()).RowsAffected == 1
}

func (m *HistoryModel) records(serviceName ...string) (histories []HistoryModel, affected int64) {
	tx := DB()
	if len(serviceName) > 0 {
		tx = tx.Where("service_name IN (?)", serviceName)
	}

	if tx = tx.Find(&histories); tx.Error != nil {
		return nil, 0
	}

	return histories, tx.RowsAffected
}
