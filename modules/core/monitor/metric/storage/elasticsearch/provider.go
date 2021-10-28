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

package elasticsearch

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda/modules/core/monitor/metric"
	"github.com/erda-project/erda/modules/core/monitor/metric/storage"
	retention "github.com/erda-project/erda/modules/core/monitor/settings/retention-strategy"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/creator"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

type (
	config struct {
		WriteTimeout time.Duration `file:"write_timeout" default:"1m"`
		IndexType    string        `file:"index_type" default:"metric"`
		DummyIndex   string        `file:"dummy_index"`
	}
	provider struct {
		Cfg       *config
		Log       logs.Logger
		ES        elasticsearch.Interface `autowired:"elasticsearch"`
		Creator   creator.Interface       `autowired:"elasticsearch.index.creator@metric"`
		Retention retention.Interface     `autowired:"storage-retention-strategy@metric" optional:"true"`
		Loader    loader.Interface        `autowired:"elasticsearch.index.loader@metric" optional:"true"`
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	if p.Retention != nil {
		ctx.AddTask(func(c context.Context) error {
			p.Retention.Loading(ctx)
			return nil
		})
	}
	return p.initDummyIndex(ctx, p.ES.Client())
}

var _ storage.Storage = (*provider)(nil)

func (p *provider) NewWriter(ctx context.Context) (storekit.BatchWriter, error) {
	if p.Creator == nil || p.Retention == nil {
		return nil, fmt.Errorf("elasticsearch.index.creator@metric and storage-retention-strategy@metric is required for Writer")
	}
	w := p.ES.NewWriter(&elasticsearch.WriteOptions{
		Timeout: p.Cfg.WriteTimeout,
		Enc:     p.encodeToDocument(ctx),
	})
	return w, nil
}

func (p *provider) encodeToDocument(ctx context.Context) func(val interface{}) (index, id, typ string, body interface{}, err error) {
	return func(val interface{}) (index, id, typ string, body interface{}, err error) {
		m := val.(*metric.Metric)
		processInvalidFields(m)

		// TODO: configurable "full_cluster"
		namespace, key := "full_cluster", p.Retention.GetConfigKey(m.Name, m.Tags)
		var fixed bool
		if ttl, ok := m.Tags[MetricTagTTL]; ok && ttl == MetricTagTTLFixed {
			if docID, ok := m.Tags[MetricTagMetricID]; ok {
				id = docID
			}
			fixed = true
		} else if docID, ok := m.Tags[MetricTagMetricID]; ok {
			id = docID
			fixed = true
		}

		if fixed {
			if len(key) > 0 {
				index, err = p.Creator.FixedIndex(m.Name, namespace, key)
			} else {
				index, err = p.Creator.FixedIndex(m.Name, namespace)
			}
			if err != nil {
				return "", "", "", nil, err
			}
		} else {
			var wait <-chan error
			if len(key) > 0 {
				wait, index = p.Creator.Ensure(m.Name, namespace, key)
			} else {
				wait, index = p.Creator.Ensure(m.Name, namespace)
			}
			if wait != nil {
				select {
				case <-wait:
				case <-ctx.Done():
					return "", "", "", nil, storekit.ErrExitConsume
				}
			}
		}
		return index, id, p.Cfg.IndexType, &Document{
			Metric: m,
			Date:   getUnixMillisecond(m.Timestamp),
		}, nil
	}
}

// Document .
type Document struct {
	*metric.Metric
	Date int64 `json:"@timestamp"`
}

const maxUnixMillisecond int64 = 9999999999999

func getUnixMillisecond(ts int64) int64 {
	if ts > maxUnixMillisecond {
		return ts / int64(time.Millisecond)
	}
	return ts
}

// const value
const (
	MetricTagMetricID = "_id"
	MetricTagTTL      = "_ttl"
	MetricTagTTLFixed = "fixed"
)

const (
	esMaxValue = float64(math.MaxInt64)
	esMinValue = float64(math.MinInt64)
)

func processInvalidFields(m *metric.Metric) {
	fields := m.Fields
	if fields == nil {
		return
	}
	for k, v := range fields {
		switch val := v.(type) {
		case float64:
			if val < esMinValue || esMaxValue < val {
				fields[k] = strconv.FormatFloat(val, 'f', -1, 64)
			}
		}
	}
}

func init() {
	servicehub.Register("metric-storage-elasticsearch", &servicehub.Spec{
		Services:   []string{"metric-storage-writer"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
