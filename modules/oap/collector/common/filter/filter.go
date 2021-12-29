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

package filter

type Config struct {
	Tagpass map[string][]string `file:"tagpass"`
}

func IsInclude(cfg Config, tags map[string]string) bool {
	if len(cfg.Tagpass) == 0 {
		return true
	}
	for k, list := range cfg.Tagpass {
		val, ok := tags[k]
		if ok {
			for _, vv := range list {
				if vv == val {
					return true
				}
			}
		}
	}
	return false
}
