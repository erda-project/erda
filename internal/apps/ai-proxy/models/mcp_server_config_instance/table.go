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
	"encoding/json"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server_config_instance/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/sirupsen/logrus"
)

type McpServerConfigInstance struct {
	common.BaseModel
	InstanceName string `gorm:"column:instance_name;type:varchar(255)"`
	ClientID     string `gorm:"column:client_id;type:char(36);not null" json:"client_id"`
	Config       string `gorm:"column:config;type:text" json:"config"`
	McpName      string `gorm:"column:mcp_name;type:varchar(255);not null" json:"mcp_name"`
	Version      string `gorm:"column:version;type:varchar(128);not null" json:"version"`
}

// TableName 指定表名
func (*McpServerConfigInstance) TableName() string {
	return "ai_proxy_mcp_server_config_instance"
}

func (m *McpServerConfigInstance) ToProtobuf() *pb.MCPServerConfigInstanceItem {
	config := make(map[string]*structpb.Value)
	if err := json.Unmarshal([]byte(m.Config), &config); err != nil {
		// if unmarshal failed, use empty config
		logrus.Errorf("unmarshal config failed, config:%v, err:%v", m.Config, err)
	}

	return &pb.MCPServerConfigInstanceItem{
		Config:       config,
		McpName:      m.McpName,
		Version:      m.Version,
		InstanceName: m.InstanceName,
	}
}
