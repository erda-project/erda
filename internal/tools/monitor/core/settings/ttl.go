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

package settings

import (
	"encoding/json"
	"time"
)

type ttl struct {
	TTL    int64 `json:"ttl"`
	HotTTL int64 `json:"hot_ttl"`
}

type ttlConfigMap struct {
	TTL    string `json:"ttl"`
	HotTTL string `json:"hot_ttl"`
}

func (t *ttl) MarshalJSON() ([]byte, error) {
	res := ttlConfigMap{
		TTL:    time.Duration(t.TTL * 24 * int64(time.Hour)).String(),
		HotTTL: time.Duration(t.HotTTL * 24 * int64(time.Hour)).String(),
	}

	return json.Marshal(&res)
}
