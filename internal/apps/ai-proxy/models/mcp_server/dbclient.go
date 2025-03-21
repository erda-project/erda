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

package mcp_server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
)

type DBClient struct {
	DB *gorm.DB
}

func (c *DBClient) CreateOrUpdate(ctx context.Context, req *pb.MCPServerRegisterRequest) (*pb.MCPServerRegisterResponse, error) {
	mcpServerConfig := &pb.MCPServerConfig{
		Tools: req.Tools,
	}

	rawConfig, err := mcpServerConfig.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mcp server config: %v", err)
	}

	var mcpServer MCPServer
	if err = c.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = tx.Where("name = ? AND version = ?", req.Name, req.Version).
			First(&mcpServer).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to query mcp server: %v", err)
			}
			// create new server
			if err = tx.Create(&MCPServer{
				ID:          uuid.New().String(),
				Name:        req.Name,
				Version:     req.Version,
				Endpoint:    req.Endpoint,
				Config:      string(rawConfig),
				IsPublished: false,
			}).Error; err != nil {
				return fmt.Errorf("failed to create mcp server: %v", err)
			}
			return nil
		}

		// update server
		mcpServer.Endpoint = req.Endpoint
		mcpServer.Description = req.Description
		mcpServer.Config = string(rawConfig)

		if err = tx.Save(&mcpServer).Error; err != nil {
			return fmt.Errorf("failed to update mcp server: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	pbMCPServer, err := mcpServer.ToProtobuf()
	if err != nil {
		return nil, err
	}

	return &pb.MCPServerRegisterResponse{
		Data: pbMCPServer,
	}, nil
}

func (c *DBClient) Publish(ctx context.Context, req *pb.MCPServerActionPublishRequest) (*pb.MCPServerActionPublishResponse, error) {
	if err := c.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var server MCPServer
		if err := tx.Where("name = ? AND version = ?", req.Name, req.Version).
			First(&server).Error; err != nil {
			return fmt.Errorf("failed to query mcp server: %v", err)
		}

		switch req.Action {
		case pb.MCPServerActionPublishType_PUT_ON:
			server.IsPublished = true
		case pb.MCPServerActionPublishType_PUT_OFF:
			server.IsPublished = false
		default:
			return fmt.Errorf("invalid action: %v", req.Action)
		}

		if err := tx.Save(&server).Error; err != nil {
			return fmt.Errorf("failed to update mcp server publish status: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &pb.MCPServerActionPublishResponse{}, nil
}

func (c *DBClient) Get(ctx context.Context, req *pb.MCPServerGetRequest) (*pb.MCPServerGetResponse, error) {
	tx := c.DB.WithContext(ctx).Where("name = ? and is_published = ?", req.Name, true)
	if req.Version != "" {
		tx = tx.Where("version = ?", req.Version)
	}

	var server MCPServer
	if err := tx.First(&server).Error; err != nil {
		return nil, fmt.Errorf("failed to query mcp server: %v", err)
	}

	pbMCPServer, err := server.ToProtobuf()
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp server to protobuf: %v", err)
	}

	return &pb.MCPServerGetResponse{
		Data: pbMCPServer,
	}, nil
}

func (c *DBClient) Paging(ctx context.Context, req *pb.MCPServerListRequest) (*pb.MCPServerListResponse, error) {
	tx := c.DB.Model(&MCPServer{}).WithContext(ctx)
	if req.Name != "" {
		tx = tx.Where("name = ?", req.Name)
	}

	if !req.IncludeUnpublished {
		tx = tx.Where("is_published = ?", true)
	}

	var (
		total int64
		list  MCPServers
	)
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.PageNum - 1) * req.PageSize
	err := tx.Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error
	if err != nil {
		return nil, err
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count mcp servers: %v", err)
	}

	servers, err := list.ToProtobuf()
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp servers to protobuf: %v", err)
	}

	return &pb.MCPServerListResponse{
		Total: total,
		List:  servers,
	}, nil
}
