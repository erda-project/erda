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

package tagger

import (
	"github.com/erda-project/erda/modules/oap/collector/core/model"
)

func (p *provider) processMetrics(metrics *model.Metrics) (model.ObservableData, error) {
	pcfg := p.Cfg.Matchers.Pod
	if pcfg == nil {
		return metrics, nil
	}
	// 1. filter
	// 2. find
	// 3. tagger
	for _, item := range metrics.Metrics {
		// filter
		for _, f := range pcfg.Filters {
			if item.Attributes[f.Key] != f.Value {
				break
			}
		}

		// find
		meta, ok := p.podCache.Get(item.Attributes[pcfg.Finder.NameKey], item.Attributes[pcfg.Finder.NamespaceKey])
		if !ok {
			continue
		}
		// tagger
		mergeMap(item.Attributes, meta)
	}
	return metrics, nil
}

func mergeMap(dst, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}
