// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mcp_filesystem

import (
	"time"
)

type BucketDomainRelation struct {
	ID        string     `gorm:"type:char(36);primaryKey;comment:primary key" json:"id"`
	CreatedAt time.Time  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt *time.Time `gorm:"type:datetime;comment:删除时间" json:"deleted_at" gorm:"index"`

	BucketName  string `gorm:"type:varchar(64);not null;comment:oss bucket name" json:"bucket_name"`
	Domain      string `gorm:"type:varchar(1024);not null;comment:oss bucket domain" json:"domain"`
	Area        string `gorm:"type:varchar(128);not null;comment:bucket area" json:"area"`
	StorageType string `gorm:"type:varchar(32);not null;comment:存储类型，如：oss, s3" json:"storage_type"`
}

// TableName sets the insert table name for this struct type
func (*BucketDomainRelation) TableName() string {
	return "ai_proxy_bucket_domain_relation"
}

type McpFile struct {
	ID        string     `gorm:"type:char(36);primaryKey;comment:primary key" json:"id"`
	CreatedAt time.Time  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt *time.Time `gorm:"type:datetime;comment:删除时间" json:"deleted_at" gorm:"index"`

	StorageType string `gorm:"type:varchar(32);not null;comment:存储类型，如：oss, s3" json:"storage_type"`
	ObjectKey   string `gorm:"type:varchar(512);not null;comment:存储对象 key" json:"object_key"`
	FileName    string `gorm:"type:varchar(512);not null;comment:文件名称" json:"file_name"`
	FileSize    int64  `gorm:"type:bigint;not null;comment:文件大小，单位：字节" json:"file_size"`
	FileMd5     string `gorm:"type:varchar(32);not null;comment:文件 md5 值" json:"file_md5"`
	VersionID   string `gorm:"type:varchar(128);not null;comment:文件版本id" json:"version_id"`
	Keep        string `gorm:"type:char(1);not null;comment:是否保留文件" json:"keep"`
	ETag        string `gorm:"type:varchar(512);not null;comment:文件 etag" json:"e_tag"`
	IsDeleted   string `gorm:"type:char(1);not null;comment:是否删除" json:"is_deleted"`
	RelationId  string `gorm:"type:char(36);not null;comment:bucket关联id" json:"relation_id"`
}

// TableName sets the insert table name for this struct type
func (*McpFile) TableName() string {
	return "ai_proxy_mcp_files"
}
