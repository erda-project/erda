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

package indexmanager

import (
	"context"
	"time"

	indexloader "github.com/erda-project/erda/modules/core/monitor/metric/index-loader"
)

func (p *provider) runIndexRollover(ctx context.Context) {
	p.Loader.WaitAndGetIndices(ctx) // Let indices load first
	p.Log.Infof("enable index rollover with interval(%v)", p.Cfg.RolloverInterval)
	timer := time.NewTimer(10 * time.Second) // 10s, to avoid restart in short time
	for {
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}
		p.RolloverIndices(ctx, func(*indexloader.IndexEntry) bool { return true })
		timer.Reset(p.Cfg.RolloverInterval)
	}
}

// RolloverIndices .
func (p *provider) RolloverIndices(ctx context.Context, filter IndexMatcher) error {
	return p.rolloverIndices(filter, p.rolloverBody)
}

func (p *provider) rolloverIndices(filter IndexMatcher, body string) error {
	indices := p.Loader.AllIndices()
	if len(indices) <= 0 {
		return nil
	}
	var num int
	for metric, mg := range indices {
		for ns, ng := range mg.Groups {
			if len(ng.List) > 0 && ng.List[0].Num > 0 && filter(ng.List[0]) {
				alias := p.indexAlias(metric, ns)
				ok, _ := p.rolloverAlias(alias, body)
				if ok {
					num++
				}
			}
			for key, kg := range ng.Groups {
				if len(kg.List) > 0 && kg.List[0].Num > 0 && filter(kg.List[0]) {
					alias := p.indexAlias(metric, ns+"."+key)
					ok, _ := p.rolloverAlias(alias, body)
					if ok {
						num++
					}
				}
			}
		}
	}
	if num > 0 {
		p.Loader.ReloadIndices()
	}
	return nil
}

func (p *provider) rolloverAlias(alias, body string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.RequestTimeout)
	defer cancel()
	resp, err := p.ES.Client().RolloverIndex(alias).BodyString(body).Do(ctx)
	if err != nil {
		p.Log.Errorf("failed to rollover alias %s : %s", alias, err)
		return false, err
	}
	if resp.Acknowledged {
		p.Log.Infof("rollover alias %s from %s to %s, %v", alias, resp.OldIndex, resp.NewIndex, resp.Acknowledged)
	}
	return resp.Acknowledged, nil
}

func (p *provider) indexAlias(name, suffix string) string {
	return p.indexPrefix + "-" + name + "-" + suffix + "-rollover"
}
