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

package model

import (
	"fmt"

	"github.com/erda-project/erda/modules/oap/collector/common/filter"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
)

// semantic same as https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#metric-filtering
// key* <=> tag*
// TODO. infra's config parser don't supported embed config
type FilterConfig struct {
	// Selectors
	Keypass    map[string][]string `file:"keypass"`
	Keydrop    map[string][]string `file:"keydrop"`
	Keyinclude []string            `file:"keyinclude"`
	Keyexclude []string            `file:"keyexclude"`
}

type DataFilter struct {
	Keypass    map[string]filter.Filter
	Keydrop    map[string]filter.Filter
	Keyinclude []string
	Keyexclude []string
}

// https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#selectors
func (df *DataFilter) Selected(od odata.ObservableData) bool {
	attr := od.Pairs()
	if df.Keypass != nil {
		for k, subf := range df.Keypass {
			val, ok := attr[k]
			if !ok {
				continue
			}
			sval, ok := val.(string)
			if !ok {
				continue
			}
			if !subf.Match(sval) {
				return false
			}
		}
	}

	if df.Keydrop != nil {
		for k, subf := range df.Keydrop {
			val, ok := attr[k]
			if !ok {
				continue
			}
			sval, ok := val.(string)
			if !ok {
				continue
			}
			if subf.Match(sval) {
				return false
			}
		}
	}

	for _, key := range df.Keyinclude {
		_, ok := attr[key]
		if !ok {
			return false
		}
	}
	for _, key := range df.Keyexclude {
		_, ok := attr[key]
		if ok {
			return false
		}
	}
	return true
}

func NewDataFilter(cfg FilterConfig) (*DataFilter, error) {
	keypass := make(map[string]filter.Filter)
	for k, v := range cfg.Keypass {
		tmp, err := filter.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("keypass<%s>: %w", k, err)
		}
		keypass[k] = tmp
	}
	keydrop := make(map[string]filter.Filter)
	for k, v := range cfg.Keydrop {
		tmp, err := filter.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("keydrop<%s>: %w", k, err)
		}
		keydrop[k] = tmp
	}

	return &DataFilter{
		Keypass:    keypass,
		Keydrop:    keydrop,
		Keyinclude: cfg.Keyinclude,
		Keyexclude: cfg.Keyexclude,
	}, nil
}
