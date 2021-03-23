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

func transformStrSliceToKeys(decrypt bool, keys ...string) []apistructs.PipelineCmsConfigKey {
	result := make([]apistructs.PipelineCmsConfigKey, 0, len(keys))
	for _, key := range keys {
		result = append(result, apistructs.PipelineCmsConfigKey{
			Key:     key,
			Decrypt: decrypt,
		})
	}
	return result
}
