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

package analysis

import (
	"github.com/recallsong/go-utils/encoding"
	"github.com/recallsong/go-utils/reflectx"

	"github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors"
	_ "github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors/regex" //
)

type processorConfig struct {
	Type   string            `json:"type"`
	Config encoding.RawBytes `json:"config"`
}

type tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (p *provider) loadProcessors() error {
	list, err := p.db.LogMetricConfig.QueryEnabledByScope(p.C.Processors.Scope, p.C.Processors.ScopeID)
	if err != nil {
		return err
	}
	ps := processors.New()
	for _, item := range list {
		if len(item.Filters) <= 0 {
			continue
		}
		var taglist []*tag
		err := json.Unmarshal(reflectx.StringToBytes(item.Filters), &taglist)
		if err != nil {
			p.L.Debugf("fail to parse log filters: %s", err)
			continue
		}
		tags := make(map[string]string, len(taglist)+4)
		for _, item := range taglist {
			tags[item.Key] = item.Value
		}
		var configs []*processorConfig
		err = json.Unmarshal(reflectx.StringToBytes(item.Processors), &configs)
		if err != nil {
			p.L.Debugf("fail to parse log processors: %s", err)
			continue
		}
		for _, cfg := range configs {
			err := ps.Add(item.ScopeID, tags, item.Metric, cfg.Type, cfg.Config)
			if err != nil {
				p.L.Debugf("fail to add log processor: %s", err)
				continue
			}
		}
	}
	p.processors.Store(ps)
	return nil
}
