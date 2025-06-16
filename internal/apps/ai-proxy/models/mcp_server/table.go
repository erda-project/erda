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
	"encoding/json"
	"time"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
)

type MCPServer struct {
	ID               string     `gorm:"type:char(36);primary_key;not null"`
	Name             string     `gorm:"type:varchar(64);not null"`
	Version          string     `gorm:"type:varchar(64);not null"`
	Description      string     `gorm:"type:text"`
	Instruction      string     `gorm:"type:text"`
	Endpoint         string     `gorm:"type:varchar(191);not null"`
	TransportType    string     `gorm:"type:varchar(64);not null"`
	Config           string     `gorm:"type:text;not null"`
	ServerConfig     string     `gorm:"type:server_config;not null;default:''"`
	IsPublished      bool       `gorm:"type:boolean;not null;default:false"`
	IsDefaultVersion bool       `gorm:"type:boolean;not null;default:false"`
	CreatedAt        time.Time  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time  `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" sql:"on update CURRENT_TIMESTAMP"`
	DeletedAt        *time.Time `gorm:"type:datetime;default:null"`
}

func (*MCPServer) TableName() string { return "ai_proxy_mcp_server" }

func (m *MCPServer) ToProtobuf() (*pb.MCPServer, error) {
	var mcpServerConfig *pb.MCPServerConfig
	if err := json.Unmarshal([]byte(m.Config), &mcpServerConfig); err != nil {
		return nil, err
	}

	return &pb.MCPServer{
		Id:               m.ID,
		Name:             m.Name,
		Version:          m.Version,
		Description:      m.Description,
		Instruction:      m.Instruction,
		Endpoint:         m.Endpoint,
		TransportType:    m.TransportType,
		Tools:            mcpServerConfig.Tools,
		ServerConfig:     m.ServerConfig,
		IsPublished:      m.IsPublished,
		IsDefaultVersion: m.IsDefaultVersion,
	}, nil
}

type MCPServers []*MCPServer

func (models MCPServers) ToProtobuf() ([]*pb.MCPServer, error) {
	var pbClients []*pb.MCPServer
	for _, c := range models {
		pbMcpServer, err := c.ToProtobuf()
		if err != nil {
			return nil, err
		}
		pbClients = append(pbClients, pbMcpServer)
	}

	return pbClients, nil
}
