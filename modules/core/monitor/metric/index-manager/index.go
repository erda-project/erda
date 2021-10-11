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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/olivere/elastic"

	"github.com/erda-project/erda/modules/core/monitor/metric"
	"github.com/erda-project/erda/pkg/router"
)

const timeForSplitIndex int64 = 24 * int64(time.Hour)

// GetWriteIndex .
// if rollover is enable, return index alias with format:
// spot-<metric>-<namespace>-r-rollover
// spot-<metric>-<namespace>.<key>-r-rollover
//
// if rollover is disable, return index with format:
// spot-<metric>-<namespace>-<timestamp>
// spot-<metric>-<namespace>.<key>-<timestamp>
func (p *provider) GetWriteIndex(metric *metric.Metric) (string, bool) {
	ns, key := p.getNamespace(metric), p.getKey(metric)
	suffix := p.getIndexSuffix(ns, key)
	name := normalizeIndexSegmentName(strings.ToLower(metric.Name))

	if p.Cfg.EnableRollover {
		alias := p.indexAlias(name, suffix)
		// check if the index exists
		indices := p.Loader.WaitAndGetIndices(context.Background())
		var find bool
		metricGroup, ok := indices[name]
		if ok {
			nsGroup, ok := metricGroup.Groups[ns]
			if ok {
				if len(key) > 0 {
					keysGroup, ok := nsGroup.Groups[key]
					if ok {
						if len(keysGroup.List) > 0 && keysGroup.List[0].Num > 0 {
							find = true
						}
					}
				} else {
					if len(nsGroup.List) > 0 && nsGroup.List[0].Num > 0 {
						find = true
					}
				}
			}
		}
		return alias, find
	}
	timestamp := (metric.Timestamp - metric.Timestamp%timeForSplitIndex) / 1000000
	return p.indexPrefix + "-" + name + "-" + suffix + "-" + strconv.FormatInt(timestamp, 10), true
}

// GetWriteFixedIndex return index with format spot-<metric>-<namespace>
func (p *provider) GetWriteFixedIndex(metric *metric.Metric) string {
	return p.indexPrefix + "-" + normalizeIndexSegmentName(strings.ToLower(metric.Name)) + "-" +
		p.getIndexSuffix(p.getNamespace(metric), p.getKey(metric))
}

func (p *provider) getNamespace(metric *metric.Metric) string {
	ns := p.namespaces.Find(metric.Name, metric.Tags)
	if ns != nil {
		return ns.(string)
	}
	return p.defaultNamespace
}

func (p *provider) getIndexSuffix(ns, key string) string {
	if len(key) > 0 {
		return ns + "." + key
	}
	return ns
}

// CreateIndex .
func (p *provider) CreateIndex(metric *metric.Metric) error {
	ns, key := p.getNamespace(metric), p.getKey(metric)
	suffix := p.getIndexSuffix(ns, key)
	name := normalizeIndexSegmentName(strings.ToLower(metric.Name))
	alias := p.indexAlias(name, suffix)

	p.createdLock.Lock()
	defer p.createdLock.Unlock()
	if p.created[alias] {
		return nil // avoid duplicate index creation
	}
	index := p.indexPrefix + "-" + name + "-" + suffix + "-r-000001"
	err := p.createIndexWithRetry(
		index, // first index
		alias,
	)
	if err != nil {
		p.Log.Error(err)
		return err
	}
	p.created[alias] = true
	p.Log.Infof("create index %q with alias %q ok", index, alias)
	return nil
}

func (p *provider) createIndexWithRetry(index, alias string) (err error) {
	createIndex := func(index, alias string) (*elastic.IndicesCreateResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.RequestTimeout)
		defer cancel()
		return p.ES.Client().CreateIndex(index).BodyJson(
			map[string]interface{}{
				"aliases": map[string]interface{}{
					alias: make(map[string]interface{}),
				},
			},
		).Do(ctx)
	}
	for i := 0; i < 2; i++ {
		resp, e := createIndex(index, alias)
		if e == nil {
			if resp != nil && !resp.Acknowledged {
				return fmt.Errorf("failed to create index=%q, alias=%q: not Acknowledged", index, alias)
			}
			return nil
		}
		err = e
	}
	return fmt.Errorf("failed to create index=%q, alias=%q: %s", index, alias, err)
}

func normalizeIndexSegmentName(s string) string { return strings.Replace(s, "-", "_", -1) }

func newNamespaceRouter(cfg *config) *router.Router {
	r := router.New()
	for _, item := range cfg.Namespaces {
		if len(item.Tags) == 1 && len(item.Namespace) == 0 {
			item.Namespace = item.Tags[0].Value
		}
		if len(item.Tags) <= 0 || len(item.Namespace) == 0 {
			continue
		}
		for _, name := range strings.Split(item.Name, ",") {
			r.Add(name, item.Tags, normalizeIndexSegmentName(item.Namespace))
		}
	}
	r.PrintTree(true)
	return r
}

func (p *provider) IndexType() string             { return p.Cfg.IndexType }
func (p *provider) EnableRollover() bool          { return p.Cfg.EnableRollover }
func (p *provider) IndexPrefix() string           { return p.Loader.IndexPrefix() }
func (p *provider) RequestTimeout() time.Duration { return p.Cfg.RequestTimeout }
func (p *provider) Client() *elastic.Client       { return p.ES.Client() }
