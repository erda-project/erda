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
type FilterConfig struct {
	// Selectors
	Namepass []string            `file:"namepass"`
	Namedrop []string            `file:"namedrop"`
	Tagpass  map[string][]string `file:"tagpass"`
	Tagdrop  map[string][]string `file:"tag_drop"`
	// Modifiers
	Fieldpass  []string `file:"fieldpass"`
	Fielddrop  []string `file:"fielddrop"`
	Taginclude []string `file:"taginclude"`
	Tagexclude []string `file:"tagexclude"`
}

type DataFilter struct {
	Namepass filter.Filter
	Namedrop filter.Filter
	Tagpass  map[string]filter.Filter
	Tagdrop  map[string]filter.Filter

	Fieldpass  filter.Filter
	Fielddrop  filter.Filter
	Taginclude filter.Filter
	Tagexclude filter.Filter
}

// https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#selectors
func (df *DataFilter) Selected(od odata.ObservableData) bool {
	if df.Namepass != nil && !df.Namepass.Match(od.Name()) {
		return false
	}
	if df.Namedrop != nil && df.Namedrop.Match(od.Name()) {
		return false
	}
	attr := od.Attributes()
	if df.Tagpass != nil {
		for k, subf := range df.Tagpass {
			val, ok := attr[k]
			if !ok {
				continue
			}
			if !subf.Match(val) {
				return false
			}
		}
	}

	if df.Tagdrop != nil {
		for k, subf := range df.Tagdrop {
			val, ok := attr[k]
			if !ok {
				continue
			}
			if subf.Match(val) {
				return false
			}
		}
	}
	return true
}

func NewDataFilter(cfg FilterConfig) (*DataFilter, error) {
	namepass, err := filter.Compile(cfg.Namepass)
	if err != nil {
		return nil, fmt.Errorf("namepass: %w", err)
	}
	namedrop, err := filter.Compile(cfg.Namedrop)
	if err != nil {
		return nil, fmt.Errorf("namedrop: %w", err)
	}
	tagpass := make(map[string]filter.Filter)
	for k, v := range cfg.Tagpass {
		tmp, err := filter.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("tagpass<%s>: %w", k, err)
		}
		tagpass[k] = tmp
	}
	tagdrop := make(map[string]filter.Filter)
	for k, v := range cfg.Tagdrop {
		tmp, err := filter.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("tagdrop<%s>: %w", k, err)
		}
		tagdrop[k] = tmp
	}

	fieldpass, err := filter.Compile(cfg.Fieldpass)
	if err != nil {
		return nil, fmt.Errorf("fieldpass: %w", err)
	}
	fielddrop, err := filter.Compile(cfg.Fielddrop)
	if err != nil {
		return nil, fmt.Errorf("fielddrop: %w", err)
	}
	taginclude, err := filter.Compile(cfg.Taginclude)
	if err != nil {
		return nil, fmt.Errorf("taginclude: %w", err)
	}
	tagexclude, err := filter.Compile(cfg.Tagexclude)
	if err != nil {
		return nil, fmt.Errorf("tagexclude: %w", err)
	}
	return &DataFilter{
		Namepass:   namepass,
		Namedrop:   namedrop,
		Tagpass:    tagpass,
		Tagdrop:    tagdrop,
		Fieldpass:  fieldpass,
		Fielddrop:  fielddrop,
		Taginclude: taginclude,
		Tagexclude: tagexclude,
	}, nil
}
