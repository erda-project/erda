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

package cms

import (
	"context"

	"github.com/erda-project/erda/apistructs"
)

type ConfigManager interface {
	// IdempotentCreateNS 幂等创建 ns
	IdempotentCreateNS(ctx context.Context, ns string) error
	// IdempotentDeleteNS 幂等删除 ns，ns 下的所有配置一并删除
	IdempotentDeleteNS(ctx context.Context, ns string) error
	// PrefixListNS 获取 ns 列表，支持前缀过滤
	PrefixListNS(ctx context.Context, nsPrefix string) ([]apistructs.PipelineCmsNs, error)
	// UpdateConfigs 不存在，则创建；已存在，则更新
	UpdateConfigs(ctx context.Context, ns string, kvs map[string]apistructs.PipelineCmsConfigValue) error
	// DeleteConfigs 根据 keys 删除配置。若 ns 不存在，返回空
	DeleteConfigs(ctx context.Context, ns string, keys ...string) error
	// GetConfigs 获取 ns 下配置。若 ns 不存在，则返回空。若指定 keys，则只获取指定 key 的配置
	GetConfigs(ctx context.Context, ns string, globalDecrypt bool, keys ...apistructs.PipelineCmsConfigKey) (map[string]apistructs.PipelineCmsConfigValue, error)
}

func transformKeysToStrSlice(keys ...apistructs.PipelineCmsConfigKey) []string {
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key.Key)
	}
	return result
}
