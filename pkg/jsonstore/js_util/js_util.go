package js_util

import (
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
)

// 取得 kvs 和 keys 的交集
func InterSectionKVs(kvs []storetypes.KeyValue, keys []string) ([]storetypes.KeyValue, []string) {
	interSectionKeys := []string{}
	interSectionKVs := []storetypes.KeyValue{}
	for _, key := range keys {
		for _, kv := range kvs {
			if string(kv.Key) == key {
				interSectionKeys = append(interSectionKeys, key)
				interSectionKVs = append(interSectionKVs, kv)
				break
			}
		}
	}
	return interSectionKVs, interSectionKeys
}
