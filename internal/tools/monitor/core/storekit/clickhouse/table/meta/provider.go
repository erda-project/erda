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

package meta

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/appscode/go/strings"
	"github.com/go-redis/redis"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
)

type LoadMode string

const (
	LoadFromCacheOnly      LoadMode = "LoadFromCacheOnly"
	LoadFromClickhouseOnly LoadMode = "LoadFromClickhouseOnly"
	LoadWithCache          LoadMode = "LoadWithCache"
)

type config struct {
	LoadMode       string        `file:"load_mode" default:"LoadWithCache"`
	MetaTable      string        `file:"table"`
	Database       string        `file:"database" default:"monitor"`
	ReloadInterval time.Duration `file:"reload_interval" default:"5m"`
	CacheKeyPrefix string        `file:"cache_key_prefix" default:"clickhouse-meta-loader"`
	MetaStartTime  time.Duration `file:"meta_start_time" default:"-1h"`
	IgnoreGap      time.Duration `file:"ignore_gap" default:"1h"`
	Once           bool
}

type updateMetricsRequest struct {
	Metas []MetricMeta
	Done  chan struct{}
}

type MetricMeta struct {
	Scope       string   `json:"scope,omitempty" ch:"org_name"`
	ScopeId     string   `json:"scope_id,omitempty" ch:"tenant_id"`
	MetricGroup string   `json:"metric_group,omitempty" ch:"metric_group"`
	StringKeys  []string `json:"string_keys,omitempty" ch:"sk"`
	NumberKeys  []string `json:"number_keys,omitempty" ch:"nk"`
	TagKeys     []string `json:"tag_keys,omitempty" ch:"tk"`
}

type provider struct {
	Cfg        *config
	Log        logs.Logger
	Clickhouse clickhouse.Interface `autowired:"clickhouse" inherit-label:"preferred"`
	Redis      *redis.Client        `autowired:"redis-client"`
	Election   election.Interface   `autowired:"etcd-election@meta-loader"`

	Meta            atomic.Value
	updateMetricsCh chan *updateMetricsRequest
	reloadCh        chan chan error

	loadLock              sync.Mutex
	suppressCacheLoader   bool
	needSyncTablesToCache bool
}

func (p *provider) Init(ctx servicehub.Context) error {
	if len(p.Cfg.MetaTable) == 0 {
		return fmt.Errorf("meta table is required")
	}

	switch LoadMode(p.Cfg.LoadMode) {
	case LoadFromClickhouseOnly:
		ctx.AddTask(p.runClickhouseMetaLoader, servicehub.WithTaskName("clickhouse meta loader"))
	case LoadFromCacheOnly:
		ctx.AddTask(p.runCacheMetaLoader, servicehub.WithTaskName("cache meta loader"))
	case LoadWithCache:
		p.Election.OnLeader(func(ctx context.Context) {
			p.needSyncTablesToCache = true
			_ = p.runClickhouseMetaLoader(ctx)
		})
		ctx.AddTask(p.runCacheMetaLoader, servicehub.WithTaskName("cache meta loader"))
	default:
		return fmt.Errorf("invalid load_mode")
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case req := <-p.updateMetricsCh:
			p.Meta.Store(req.Metas)
			if req.Done != nil {
				req.Done <- struct{}{}
				close(req.Done)
			}
			p.Log.Infof("load clickhouse metric, num: %d", len(req.Metas))
		}
	}
}

func (p *provider) updateMetrics(metas []MetricMeta) chan struct{} {
	ch := make(chan struct{}, 1)
	req := updateMetricsRequest{
		Metas: metas,
		Done:  ch,
	}
	p.updateMetricsCh <- &req
	return ch
}

func (meta *MetricMeta) check(scope, scopeId string) bool {
	if scope == "org" {
		return meta.Scope == scopeId || meta.Scope == ""
	}
	if scope == "micro_service" {
		return meta.ScopeId == scopeId
	}

	if meta.Scope != scope {
		return false
	}
	if !strings.IsEmpty(&scopeId) {
		return meta.ScopeId == scopeId
	}
	return true
}

func init() {
	servicehub.Register("clickhouse.meta.loader", &servicehub.Spec{
		Services: []string{"clickhouse.meta.loader"},
		Types: []reflect.Type{
			reflect.TypeOf((*Interface)(nil)).Elem(),
		},
		Dependencies:         []string{"clickhouse"},
		OptionalDependencies: []string{"etcd-election"},
		ConfigFunc:           func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			p := &provider{
				updateMetricsCh: make(chan *updateMetricsRequest, 1),
				reloadCh:        make(chan chan error),
			}
			return p
		},
	})
}
