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
