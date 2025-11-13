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

package mcp_server_config_instance

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type DBClient struct {
	db *gorm.DB
}

func NewDBClient(db *gorm.DB) *DBClient {
	return &DBClient{db}
}

func (c *DBClient) Get(ctx context.Context, mcpName, version, clientId string) (*McpServerConfigInstance, error) {
	var instance McpServerConfigInstance
	if err := c.db.WithContext(ctx).Model(&McpServerConfigInstance{}).
		Where("mcp_name = ?", mcpName).
		Where("version = ?", version).
		Where("client_id = ?", clientId).
		First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *DBClient) Count(ctx context.Context, mcpName, version, clientId string) (int64, error) {
	var count int64
	tx := c.db.WithContext(ctx).Model(&McpServerConfigInstance{}).
		Where("mcp_name = ?", mcpName).
		Where("version = ?", version)
	if clientId != "" {
		tx = tx.Where("client_id = ?", clientId)
	}

	if err := tx.Debug().Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (c *DBClient) CountAll(ctx context.Context, clientId string) ([]*McpServerConfigInstanceCountResult, error) {
	var args []any
	query := `
    SELECT mcp_name, version, COUNT(*) AS count
    FROM ai_proxy_mcp_server_config_instance`

	if clientId != "" {
		query = query + " WHERE client_id = ? "
		args = append(args, clientId)
	}

	query = query + `
    GROUP BY mcp_name, version
    ORDER BY mcp_name`

	var results []*McpServerConfigInstanceCountResult
	err := c.db.WithContext(ctx).Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (c *DBClient) CreateOrUpdate(ctx context.Context, model *McpServerConfigInstance) (*McpServerConfigInstance, error) {
	exist := true
	if err := c.db.WithContext(ctx).Model(&McpServerConfigInstance{}).Where("mcp_name = ? AND version = ? AND client_id = ?", model.McpName, model.Version, model.ClientID).First(&McpServerConfigInstance{}).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		exist = false
	}

	if exist {
		if err := c.db.WithContext(ctx).Model(&McpServerConfigInstance{}).Where("mcp_name = ? AND version = ? AND client_id = ?", model.McpName, model.Version, model.ClientID).Update("config", model.Config).Error; err != nil {
			return nil, err
		}
	} else {
		if err := c.db.WithContext(ctx).Model(&McpServerConfigInstance{}).Create(model).Error; err != nil {
			return nil, err
		}
	}

	return model, nil
}

func (c *DBClient) Delete(ctx context.Context, name, version, clientId string) error {
	tx := c.db.WithContext(ctx).Model(&McpServerConfigInstance{}).
		Where("mcp_name = ?", name).
		Where("version = ?", version).
		Where("client_id = ?", clientId)
	if err := tx.Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateConfig(ctx context.Context, name, version, clientId, config string) error {
	tx := c.db.WithContext(ctx).Model(&McpServerConfigInstance{}).
		Where("mcp_name = ?", name).
		Where("version = ?", version).
		Where("client_id = ?", clientId)

	if err := tx.Update("config", config).Error; err != nil {
		return err
	}
	return nil
}

type ListOptions struct {
	ClientId *string
	PageNum  int
	PageSize int
}

func (c *DBClient) List(ctx context.Context, options *ListOptions) ([]*McpServerConfigInstance, int64, error) {
	var (
		pageNum   = options.PageNum
		pageSize  = options.PageSize
		instances []*McpServerConfigInstance
		total     int64
	)

	tx := c.db.WithContext(ctx).Model(&McpServerConfigInstance{})

	if options.ClientId != nil {
		tx = tx.Where("client_id = ?", *options.ClientId)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pageNum > 0 && pageSize > 0 {
		tx = tx.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	}

	if err := tx.Find(&instances).Error; err != nil {
		return nil, 0, err
	}
	return instances, total, nil
}
