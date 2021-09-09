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

package addon

import (
	"encoding/json"
	"strings"
)

// IsEncryptedValueByKey Determine whether it is an encrypted field by key
func IsEncryptedValueByKey(key string) bool {
	return strings.Contains(strings.ToLower(key), "pass") || strings.Contains(strings.ToLower(key), "secret")
}

// IsEncryptedValueByValue Determine whether it is an encrypted field by value
func IsEncryptedValueByValue(value string) bool {
	return strings.Contains(value, ErdaEncryptedValue)
}

// GetAddonConfig return unmarshal config
func GetAddonConfig(cfgStr string) (map[string]interface{}, error) {
	cfg := map[string]interface{}{}
	if err := json.Unmarshal([]byte(cfgStr), &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
