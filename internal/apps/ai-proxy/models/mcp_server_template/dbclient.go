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
	"fmt"

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

type ListOption struct {
	ClientId        string
	PageSize        uint64
	PageNum         uint64
	McpName         *string
	IsExistInstance *bool
}

func (d *DBClient) List(ctx context.Context, option ListOption) ([]*McpServerTemplateWithInstanceCount, int64, error) {
	var total int64
	var pageSize uint64 = 10
	if option.PageSize != 0 {
		pageSize = option.PageSize
	}
	var pageNum uint64 = 1
	if option.PageNum != 0 {
		pageNum = option.PageNum
	}

	tx := d.db.WithContext(ctx).Model(&McpServerTemplate{})

	if err := tx.Count(&total).Error; err != nil {
		logrus.Errorf("failed to count templates, err: %v", err)
		return nil, 0, err
	}

	var results []*McpServerTemplateWithInstanceCount
	var args []any

	query := `
		SELECT
		    t.*,
		    COUNT(c.id) AS instance_count
		FROM
		    ai_proxy_mcp_server_template AS t
		LEFT JOIN
		    ai_proxy_mcp_server_config_instance AS c
		ON
		    t.mcp_name = c.mcp_name
		    AND t.version = c.version
`
	if option.ClientId != "" {
		query = query + `
		AND c.client_id = ?
`
		args = append(args, option.ClientId)
	}

	if option.McpName != nil {
		query = query + `
		WHERE t.mcp_name like ?
`
		args = append(args, fmt.Sprintf("%%%s%%", *option.McpName))
	}

	query = query + `
		GROUP BY
		    t.mcp_name, t.version
`

	// the empty template will be created automatically,
	// so if instance_count is 0, these instances need to be skipped;
	// if instance_count greater than 0, these instances need to be included
	if option.IsExistInstance != nil {
		var where = `HAVING instance_count = 0 AND t.template != '[]' AND t.template != ''`
		if *option.IsExistInstance {
			where = `HAVING instance_count > 0 OR t.template = '[]' OR t.template = ''`
		}
		query = query + where
	}

	query = query + `
		LIMIT ? OFFSET ?
`
	args = append(args, pageSize, (pageNum-1)*pageSize)

	logrus.Infof("args: %+v", args)
	if err := d.db.Debug().WithContext(ctx).Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
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
