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

package indexloader

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	"github.com/olivere/elastic"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
)

// Index .
type Interface interface {
	WaitAndGetIndices(ctx context.Context) map[string]*IndexGroup
	AllIndices() map[string]*IndexGroup
	ReloadIndices() error
	WatchLoadEvent(func(map[string]*IndexGroup))

	GetReadIndices(metrics []string, namespace []string, start, end int64) []string
	MetricNames() []string

	EmptyIndex() string
	IndexPrefix() string
	RequestTimeout() time.Duration
	QueryIndexTimeRange() bool
	Client() *elastic.Client
	URLs() string
}

type (
	config struct {
		LoadMode            string        `file:"load_mode" default:"LoadFromElasticSearchOnly"`
		RequestTimeout      time.Duration `file:"request_timeout" default:"1m" env:"METRIC_REQUEST_TIMEOUT"`
		DefaultNamespace    string        `file:"default_namespace" env:"METRIC_DEFAULT_NAMESPACE"`
		IndexPrefix         string        `file:"index_prefix" env:"METRIC_INDEX_PREFIX"`
		QueryIndexTimeRange bool          `file:"query_index_time_range"`
		IndexReloadInterval time.Duration `file:"index_reload_interval" default:"2m" env:"METRIC_INDEX_RELOAD_INTERVAL"`
		CacheKeyPrefix      string        `file:"cache_key_prefix" default:"es-index"`
	}
	provider struct {
		Cfg      *config
		Log      logs.Logger
		ES       elasticsearch.Interface `autowired:"elasticsearch"`
		Election election.Interface      `autowired:"etcd-election@index" optional:"true"`
		Redis    *redis.Client           `autowired:"redis-client" optional:"true"`
		es       *elastic.Client

		indices      atomic.Value          // map[string]*IndexGroup, loaded index
		setIndicesCh chan *indicesBundle   //
		reloadCh     chan chan error       // trigger to load index
		timeRanges   map[string]*timeRange // cache the maximum and minimum time of index
		listeners    []func(map[string]*IndexGroup)

		syncLock    sync.Mutex
		cond        sync.Cond
		inSync      bool
		startSyncCh chan struct{}
	}
	indicesBundle struct {
		indices map[string]*IndexGroup
		doneCh  chan struct{}
	}
)

const (
	loadFromCacheOnly         = "LoadFromCacheOnly"
	loadFromElasticSearchOnly = "LoadFromElasticSearchOnly"
	loadIndicesWithCache      = "LoadIndicesWithCache"
)

var _ Interface = (*provider)(nil)

func (p *provider) Init(ctx servicehub.Context) error {
	switch p.Cfg.LoadMode {
	case loadFromElasticSearchOnly:
		if p.ES == nil {
			return fmt.Errorf("elasticsearch-client is required")
		}
		ctx.AddTask(p.runElasticSearchIndexLoader, servicehub.WithTaskName("elasticsearch index loader"))
	case loadFromCacheOnly:
		if p.Redis == nil {
			return fmt.Errorf("redis-client is required")
		}
		ctx.AddTask(p.runCacheLoader, servicehub.WithTaskName("cache index loader"))
	case loadIndicesWithCache:
		if p.ES == nil || p.Redis == nil || p.Election == nil {
			return fmt.Errorf("elasticsearch-client、etcd-election、redis-client are required")
		}
		p.Election.OnLeader(p.syncIndiceToCache)
		ctx.AddTask(p.runCacheLoader, servicehub.WithTaskName("cached index loader"))
	default:
		return fmt.Errorf("invalid load_mode, only support: %v", []string{loadFromCacheOnly, loadFromElasticSearchOnly, loadIndicesWithCache})
	}
	routes := ctx.Service("http-router", interceptors.CORS()).(httpserver.Router)
	err := p.intRoutes(routes)
	if err != nil {
		return fmt.Errorf("failed to init routes: %s", err)
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case bdl := <-p.setIndicesCh:
			p.indices.Store(bdl.indices)
			if bdl.doneCh != nil {
				bdl.doneCh <- struct{}{}
			}
			for _, fn := range p.listeners {
				fn(bdl.indices)
			}
			p.Log.Infof("load indices %d, metrics: %d", p.getIndicesNum(bdl.indices), len(bdl.indices))
		}
	}
}

func (p *provider) updateIndices(bdl *indicesBundle) {
	p.setIndicesCh <- bdl
}

func (p *provider) getIndicesNum(indices map[string]*IndexGroup) (num int) {
	for _, index := range indices {
		for _, ns := range index.Groups {
			if ns.Fixed != nil {
				num++
			}
			num += len(index.List)
			for _, keys := range ns.Groups {
				if keys.Fixed != nil {
					num++
				}
				num += len(keys.List)
			}
		}
	}
	return num
}

func init() {
	servicehub.Register("erda.core.monitor.metric.index-loader", &servicehub.Spec{
		Services:     []string{"erda.core.monitor.metric.index-loader"},
		Dependencies: []string{"http-router"},
		Types: []reflect.Type{
			reflect.TypeOf((*Interface)(nil)).Elem(),
		},
		Description: "metrics index loader",
		ConfigFunc:  func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			p := &provider{
				setIndicesCh: make(chan *indicesBundle, 1),
				reloadCh:     make(chan chan error),
				startSyncCh:  make(chan struct{}),
				timeRanges:   make(map[string]*timeRange),
			}
			p.cond.L = &p.syncLock
			return p
		},
	})
}
