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

package mcp_server_template

import (
	"context"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DBClient struct {
	db *gorm.DB
}

func NewDBClient(db *gorm.DB) *DBClient {
	return &DBClient{db}
}

func (d *DBClient) Get(ctx context.Context, name string, version string) (*McpServerTemplate, error) {
	var template McpServerTemplate
	if err := d.db.WithContext(ctx).Model(&McpServerTemplate{}).
		Where("mcp_name = ? AND version = ?", name, version).
		First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (d *DBClient) List(ctx context.Context, pageSize uint64, pageNum uint64) ([]*McpServerTemplate, int64, error) {
	var templates []*McpServerTemplate
	var total int64
	if pageSize == 0 {
		pageSize = 10
	}
	if pageNum == 0 {
		pageNum = 1
	}

	tx := d.db.WithContext(ctx).Model(&McpServerTemplate{})

	if err := tx.Count(&total).Error; err != nil {
		logrus.Errorf("failed to count templates, err: %v", err)
		return nil, 0, err
	}

	if err := tx.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).Find(&templates).Error; err != nil {
		logrus.Errorf("failed to list templates, err: %v", err)
		return nil, 0, err
	}

	return templates, total, nil
}

func (d *DBClient) Create(ctx context.Context, template string, name string, version string) (*McpServerTemplate, error) {
	item := McpServerTemplate{
		Template: template,
		McpName:  name,
		Version:  version,
	}
	if err := d.db.WithContext(ctx).Model(&McpServerTemplate{}).Create(&item).Error; err != nil {
		logrus.Errorf("failed to create template, err: %v", err)
		return nil, err
	}
	return &item, nil
}
