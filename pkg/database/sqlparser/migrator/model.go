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

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const SchemaMigrationHistory = `schema_migration_history`

const CreateTableHistoryModel = `
CREATE TABLE IF NOT EXISTS schema_migration_history (
	id 			BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
	created_at 	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at 	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

	service_name	VARCHAR(32) 	NOT NULL COMMENT '微服务名',
	filename		VARCHAR(191) 	NOT NULL COMMENT '脚本文件名',
	checksum		CHAR(128)		NOT NULL COMMENT '脚本文本 checksum',
	installed_by	VARCHAR(32)		NOT NULL COMMENT '执行人',
	installed_on	VARCHAR(32)		NOT NULL COMMENT '执行平台',
	language_type	VARCHAR(16)		NOT NULL COMMENT '.sql, .py',
	reversed 		LONGTEXT 		NOT NULL COMMENT '反转的 DDL',
	CONSTRAINT uk_filename UNIQUE (filename),
	CONSTRAINT uk_checksum UNIQUE (checksum)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'erda migration 执行记录表'
`

type HistoryModel struct {
	ID        uint64    `gorm:"primary key"`
	CreatedAt time.Time `json:"createdAt" gorm:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"updated_at"`

	ServiceName  string `json:"serviceName"  gorm:"service_name"`
	Filename     string `json:"filename"     gorm:"filename"`
	Checksum     string `json:"checksum"     gorm:"checksum"`
	InstalledBy  string `json:"installedBy"  gorm:"installed_by"`
	InstalledOn  string `json:"installedOn"  gorm:"install_on"`
	LanguageType string `json:"languageType" gorm:"language_type"`
	Reversed     string `json:"reversed"     gorm:"reversed"`
}

func (m HistoryModel) TableName() string {
	return SchemaMigrationHistory
}

func (m HistoryModel) CreateTable(db *gorm.DB) {
	db.Logger.LogMode(logger.Silent)
	if err := db.Exec(CreateTableHistoryModel).Error; err != nil {
		logrus.WithError(err).Fatal("failed to create table")
	}
}
