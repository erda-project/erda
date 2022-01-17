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

import (
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// semantic same as https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#metric-filtering
type Config struct {
	Namepass  []string            `file:"namepass"`
	Tagpass   map[string][]string `file:"tagpass"`
	Fieldpass []string            `file:"fieldpass"`
}

func (cfg Config) IsPass(item *model.DataItem) bool {
	return cfg.IsTagpass(item.Tags) && cfg.IsFieldpass(item.Fields) && cfg.IsNamepass(item.Name)
}

// IsTagpass.
func (cfg Config) IsTagpass(tags map[string]string) bool {
	if len(cfg.Tagpass) == 0 {
		return true
	}
	for k, list := range cfg.Tagpass {
		val, ok := tags[k]
		if !ok {
			continue
		}
		for _, vv := range list {
			if vv == val {
				return true
			}
		}
	}
	return false
}

// IsFieldpass.
func (cfg Config) IsFieldpass(fields map[string]*structpb.Value) bool {
	if len(cfg.Fieldpass) == 0 {
		return true
	}
	for _, key := range cfg.Fieldpass {
		_, ok := fields[key]
		if ok {
			return true
		}
	}
	return false
}

func (cfg Config) IsNamepass(name string) bool {
	if len(cfg.Namepass) == 0 {
		return true
	}
	for _, key := range cfg.Namepass {
		if key == name {
			return true
		}
	}
	return false
}
