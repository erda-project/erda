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

package client_mcp_relation

import "github.com/erda-project/erda/internal/apps/ai-proxy/models/common"

type ClientMcpRelation struct {
	common.BaseModel
	ClientID  string `gorm:"column:client_id;type:char(36);not null;comment:client id;uniqueIndex:unique_clientid_modelid"`
	ScopeType string `gorm:"column:scope_type;type:char(36);not null;comment:MCP scope;uniqueIndex:unique_clientid_modelid"`
	ScopeID   string `gorm:"column:scope_id;type:char(36);not null;comment:MCP scope id;uniqueIndex:unique_clientid_modelid"`
}

// TableName 覆盖表名
func (*ClientMcpRelation) TableName() string {
	return "ai_proxy_client_mcp_relation"
}
