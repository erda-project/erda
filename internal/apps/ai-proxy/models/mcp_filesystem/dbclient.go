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

import "gorm.io/gorm"

type DBClient struct {
	DB *gorm.DB
}

func (d *DBClient) CreateOrUpdateBucketDomainRelation(relation BucketDomainRelation) error {
	return d.DB.Model(&BucketDomainRelation{}).Save(relation).Error
}

func (d *DBClient) GetRelationByBucket(bucket string) (*BucketDomainRelation, error) {
	var relation BucketDomainRelation
	if err := d.DB.Model(&BucketDomainRelation{}).Where("bucket_name = ?", bucket).First(&relation).Error; err != nil {
		return nil, err
	}
	return &relation, nil
}

// InsertFile McpFile
func (d *DBClient) InsertFile(file McpFile) error {
	return d.DB.Model(&McpFile{}).Create(file).Error
}

func (d *DBClient) ListMcpFiles() ([]*McpFile, error) {
	var list []*McpFile
	if err := d.DB.Model(&McpFile{}).Where("is_deleted = 'N'").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (d *DBClient) ListMcpFilesNoNeedKeep() ([]*McpFile, error) {
	var list []*McpFile
	if err := d.DB.Model(&McpFile{}).
		Where("keep != ?", "Y").
		Where("is_deleted = 'N'").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (d *DBClient) GetFileById(id string) (*McpFile, error) {
	var file McpFile
	if err := d.DB.Model(&McpFile{}).
		Where("id = ?", id).
		Where("is_deleted = 'N'").
		First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (d *DBClient) GetFileByObjectKey(key string) (*McpFile, error) {
	var file McpFile
	if err := d.DB.Model(&McpFile{}).
		Where("object_key = ?", key).
		Where("is_deleted = 'N'").
		First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (d *DBClient) DeleteFile(id string) error {
	return d.DB.Model(&McpFile{}).Where("id = ?", id).Update("is_deleted", "Y").Error
}

func (d *DBClient) DeleteFileByKey(key string) error {
	return d.DB.Model(&McpFile{}).Where("object_key = ?", key).Update("is_deleted", "Y").Error
}
