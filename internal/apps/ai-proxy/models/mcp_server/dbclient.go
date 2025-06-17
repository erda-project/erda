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
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
)

type DBClient struct {
	DB *gorm.DB
}

type ListOptions struct {
	PageNum            int
	PageSize           int
	Name               string
	IncludeUnpublished bool
}

func (c *DBClient) CreateOrUpdate(ctx context.Context, req *pb.MCPServerRegisterRequest) (*pb.MCPServerRegisterResponse, error) {
	mcpServerConfig := &pb.MCPServerConfig{
		Tools: req.Tools,
	}

	transportType := req.TransportType
	if transportType == "" {
		transportType = "sse"
	}

	rawConfig, err := mcpServerConfig.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mcp server config: %v", err)
	}

	var dbServer MCPServer
	if err = c.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = tx.Where("name = ? AND version = ?", req.Name, req.Version).
			First(&dbServer).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to query mcp server: %v", err)
			}

			dbServer = MCPServer{
				ID:               uuid.New().String(),
				Name:             req.Name,
				Description:      req.Description,
				Instruction:      req.Instruction,
				Version:          req.Version,
				Endpoint:         req.Endpoint,
				TransportType:    transportType,
				Config:           string(rawConfig),
				ServerConfig:     req.ServerConfig,
				IsPublished:      req.IsPublished != nil && req.IsPublished.Value,
				IsDefaultVersion: req.IsDefaultVersion != nil && req.IsDefaultVersion.Value,
			}

			// create new server
			if err = tx.Create(&dbServer).Error; err != nil {
				return fmt.Errorf("failed to create mcp server: %v", err)
			}

			// set current version to default.
			if dbServer.IsDefaultVersion {
				if err = tx.Model(&MCPServer{}).
					Where("name = ? and version != ?", req.Name, req.Version).
					Update("is_default_version", false).Error; err != nil {
					return fmt.Errorf("failed to update mcp server: %v", err)
				}
			}
			return nil
		}

		// update server
		dbServer.Endpoint = req.Endpoint
		dbServer.TransportType = transportType
		dbServer.Description = req.Description
		dbServer.Instruction = req.Instruction
		dbServer.Config = string(rawConfig)
		dbServer.ServerConfig = req.ServerConfig
		dbServer.IsPublished = req.IsPublished != nil && req.IsPublished.Value

		// set current version to default.
		if dbServer.IsDefaultVersion {
			if err = tx.Model(&MCPServer{}).
				Where("name = ? and version != ?", req.Name, req.Version).
				Update("is_default_version", false).Error; err != nil {
				return fmt.Errorf("failed to update mcp server: %v", err)
			}
		}

		if err = tx.Save(&dbServer).Error; err != nil {
			return fmt.Errorf("failed to update mcp server: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	server, err := dbServer.ToProtobuf()
	if err != nil {
		return nil, err
	}

	return &pb.MCPServerRegisterResponse{
		Data: server,
	}, nil
}

func (c *DBClient) Publish(ctx context.Context, req *pb.MCPServerActionPublishRequest) (*pb.MCPServerActionPublishResponse, error) {
	if err := c.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dbServer MCPServer
		if err := tx.Where("name = ? AND version = ?", req.Name, req.Version).
			First(&dbServer).Error; err != nil {
			return fmt.Errorf("failed to query mcp server: %v", err)
		}

		switch req.Action {
		case pb.MCPServerActionPublishType_PUT_ON:
			dbServer.IsPublished = true
		case pb.MCPServerActionPublishType_PUT_OFF:
			dbServer.IsPublished = false
		default:
			return fmt.Errorf("invalid action: %v", req.Action)
		}

		if err := tx.Save(&dbServer).Error; err != nil {
			return fmt.Errorf("failed to update mcp dbServer publish status: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &pb.MCPServerActionPublishResponse{}, nil
}

func (c *DBClient) Get(ctx context.Context, req *pb.MCPServerGetRequest) (*pb.MCPServerGetResponse, error) {
	tx := c.DB.WithContext(ctx).
		Where("name = ? and is_published = ?", req.Name, true)

	if req.Version != "" {
		tx = tx.Where("version = ?", req.Version)
	} else {
		tx = tx.Where("is_default_version = ?", true)
	}

	var dbMCPServer MCPServer
	if err := tx.First(&dbMCPServer).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to query mcp server: %v", err)
		}

		var versions []string
		if err := c.DB.WithContext(ctx).Model(&MCPServer{}).
			Select("version").
			Where("name = ? and is_published = ?", req.Name, true).
			Order("created_at").
			Find(&versions).Error; err != nil {
			return nil, fmt.Errorf("failed to query mcp server versions: %v", err)
		}

		if req.Version == "" {
			return nil, fmt.Errorf("mcp server %s default version not found, support version: [%s]",
				req.Name, strings.Join(versions, ","))
		}

		return nil, fmt.Errorf("mcp server %s not found, supported version: [%s]",
			req.Name, strings.Join(versions, ","))
	}

	server, err := dbMCPServer.ToProtobuf()
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp server to protobuf: %v", err)
	}

	return &pb.MCPServerGetResponse{
		Data: server,
	}, nil
}

func (c *DBClient) List(ctx context.Context, options *ListOptions) (int64, []*pb.MCPServer, error) {
	var (
		total int64
		list  MCPServers
	)
	if options.PageNum == 0 {
		options.PageNum = 1
	}
	if options.PageSize == 0 {
		options.PageSize = 20
	}

	tx := c.DB.WithContext(ctx).Model(&MCPServer{})
	if options.Name != "" {
		tx = tx.Where("name = ?", options.Name)
	}

	if !options.IncludeUnpublished {
		tx = tx.Where("is_published = ?", true)
	}

	offset := (options.PageNum - 1) * options.PageSize
	err := tx.Order("created_at DESC").Limit(options.PageSize).Offset(offset).Find(&list).Error
	if err != nil {
		return 0, nil, err
	}

	if err = tx.Count(&total).Error; err != nil {
		return 0, nil, fmt.Errorf("failed to count mcp servers: %v", err)
	}

	servers, err := list.ToProtobuf()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to convert mcp servers to protobuf: %v", err)
	}

	return total, servers, nil
}

func (c *DBClient) Delete(ctx context.Context, req *pb.MCPServerDeleteRequest) (*pb.MCPServerDeleteResponse, error) {
	tx := c.DB.WithContext(ctx).Where("name = ?", req.Name)
	if req.Version != "" {
		tx = tx.Where("version = ?", req.Version)
	}
	if err := tx.Delete(&MCPServer{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete mcp server: %v", err)
	}
	return &pb.MCPServerDeleteResponse{}, nil
}

func (c *DBClient) Update(ctx context.Context, req *pb.MCPServerUpdateRequest) (*pb.MCPServerUpdateResponse, error) {
	var dbMcpServer MCPServer

	if err := c.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("name = ? and version = ?", req.Name, req.Version).
			First(&dbMcpServer).Error; err != nil {
			return err
		}

		if req.Description != "" {
			dbMcpServer.Description = req.Description
		}

		if req.Instruction != "" {
			dbMcpServer.Instruction = req.Instruction
		}

		if req.IsPublished != nil {
			dbMcpServer.IsPublished = req.IsPublished.Value
		}

		if req.IsDefaultVersion != nil {
			dbMcpServer.IsDefaultVersion = req.IsDefaultVersion.Value
		}

		if dbMcpServer.IsDefaultVersion {
			if err := tx.Model(&MCPServer{}).
				Where("name = ? and version != ?", req.Name, req.Version).
				Update("is_default_version", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(&dbMcpServer).Error
	}); err != nil {
		return nil, err
	}

	pbServer, err := dbMcpServer.ToProtobuf()
	if err != nil {
		return nil, err
	}

	return &pb.MCPServerUpdateResponse{
		Data: pbServer,
	}, nil
}
