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
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server_template/pb"
)

const EmptyTemplate = "[]"

const MCPErdaProvider = "Erda Platform"
const MCPUnknowProvider = "Unknow Provider"

const TemplateItemScopeHeader = "header"
const TemplateItemScopeQuery = "query"
const TemplateItemScopeNotification = "notification"

type TemplateItem struct {
	Description string  `json:"description"`
	Name        string  `json:"name"`
	Required    bool    `json:"required"`
	Type        string  `json:"type"`
	Scope       *string `json:"scope,omitempty"`
	Default     *string `json:"default,omitempty"`
}

type McpServerTemplate struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	McpName     string `gorm:"column:mcp_name;type:varchar(255);not null" json:"mcp_name"`
	Version     string `gorm:"column:version;type:varchar(128);not null" json:"version"`
	Template    string `gorm:"column:template;type:text" json:"template"`
	Description string `gorm:"column:description;type:text" json:"description"`
	ScopeType   string `gorm:"column:scope_type;type:varchar(64);not null" json:"scope_type"`
	ScopeID     string `gorm:"column:scope_id;type:varchar(64);not null" json:"scope_id"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;not null" json:"updated_at"`
	DeletedAt time.Time `gorm:"column:deleted_at;default:'1970-01-01 00:00:00'" json:"deleted_at"`
}

func (*McpServerTemplate) TableName() string {
	return "ai_proxy_mcp_server_template"
}

func (m *McpServerTemplate) ToProtobuf() *pb.MCPServerTemplateItem {
	var template = make([]*pb.TemplateItem, 0)

	if m.Template != "" && m.Template != "[]" {
		if err := json.Unmarshal([]byte(m.Template), &template); err != nil {
			// if it is error, just set template is [], no need to return
			logrus.Errorf("failed to unmarshal template, err: %v, %s", err, m.Template)
		}
	}

	// TODO: Abstract the MCP provider, temporarily set the scopeId with value 0 to `Erda Platform`.
	var provider = MCPUnknowProvider
	if m.ScopeType == "platform" && m.ScopeID == "0" {
		provider = MCPErdaProvider
	}

	return &pb.MCPServerTemplateItem{
		Id:          int64(m.ID),
		McpName:     m.McpName,
		Version:     m.Version,
		Template:    template,
		Description: m.Description,
		Provider:    provider,
	}
}

func (m *McpServerTemplate) IsEmptyTemplate() bool {
	return m.Template == "" || m.Template == EmptyTemplate
}

type McpServerTemplateWithInstanceCount struct {
	*McpServerTemplate
	InstanceCount int64 `gorm:"column:instance_count;type:bigint(20);not null" json:"instance_count"`
}

func (m *McpServerTemplateWithInstanceCount) ToProtobuf() *pb.MCPServerTemplateItem {
	protobuf := m.McpServerTemplate.ToProtobuf()
	protobuf.InstanceCount = m.InstanceCount
	return protobuf
}
