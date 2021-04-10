// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dto

type KongCredentialReqDto struct {
	// 消费者id，必填
	ConsumerId string `json:"consumer_id"`
	// 插件名称
	PluginName string `json:"plugin_name,omitempty"`
	// 其余选填配置
	Config *KongCredentialDto `json:"config,omitempty"`
}

// IsEmpty
func (dto KongCredentialReqDto) IsEmpty() bool {
	return len(dto.ConsumerId) == 0 || len(dto.PluginName) == 0
}
