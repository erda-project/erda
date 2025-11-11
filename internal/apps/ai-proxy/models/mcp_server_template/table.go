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

type McpServerTemplate struct {
	ID       uint64 `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	McpName  string `gorm:"column:mcp_name;type:varchar(255);not null" json:"mcp_name"`
	Version  string `gorm:"column:version;type:varchar(128);not null" json:"version"`
	Template string `gorm:"column:template;type:text" json:"template"`

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
			// if it is error, just set template is {}, no need to return
			logrus.Errorf("failed to unmarshal template, err: %v", err)
		}
	}

	return &pb.MCPServerTemplateItem{
		Id:       int64(m.ID),
		McpName:  m.McpName,
		Version:  m.Version,
		Template: template,
	}
}

func (m *McpServerTemplate) IsEmptyTemplate() bool {
	return m.Template == "" || m.Template == EmptyTemplate
}
