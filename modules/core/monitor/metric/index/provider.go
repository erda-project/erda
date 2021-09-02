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
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	mutex "github.com/erda-project/erda-infra/providers/etcd-mutex"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/pkg/router"
)

type config struct {
	RequestTimeout time.Duration `file:"request_timeout" env:"METRIC_REQUEST_TIMEOUT"`

	DefaultNamespace string `file:"default_namespace" env:"METRIC_DEFAULT_NAMESPACE"`
	Namespaces       []struct {
		Name      string             `file:"name"`
		Tags      []*router.KeyValue `file:"tags"`
		Namespace string             `file:"namespace"`
	} `file:"namespaces"`

	IndexType   string `file:"index_type" env:"METRIC_INDEX_TYPE"`
	IndexPrefix string `file:"index_prefix" env:"METRIC_INDEX_PREFIX"`

	EnableIndexInit   bool   `file:"enable_index_init"`
	IndexTemplateName string `file:"index_template_name" default:"spot_metric_template"`
	IndexTemplatePath string `file:"index_template_path"`

	EnableRollover   bool          `file:"enable_rollover"`
	RolloverBodyFile string        `file:"rollover_body_file"`
	RolloverInterval time.Duration `file:"rollover_interval"`

	QueryIndexTimeRange bool          `file:"query_index_time_range"`
	IndexReloadInterval time.Duration `file:"index_reload_interval" default:"2m" env:"METRIC_INDEX_RELOAD_INTERVAL"`

	LoadIndexTTLFromDatabase bool          `file:"load_index_ttl_from_database"`
	TTLReloadInterval        time.Duration `file:"ttl_reload_interval" default:"3m"`
	EnableIndexClean         bool          `file:"enable_index_clean" env:"ENABLE_METRIC_INDEX_CLEAN"`
	IndexTTL                 time.Duration `file:"index_ttl" env:"METRIC_INDEX_TTL"`
	IndexCleanInterval       time.Duration `file:"index_clean_interval" default:"1h" env:"METRIC_INDEX_CLEAN_INTERVAL"`

	DiskClean struct {
		EnableIndexCleanByDisk bool          `file:"enable_index_clean_by_disk" env:"ENABLE_METRIC_INDEX_CLEAN_BY_DISK"`
		CheckInterval          time.Duration `file:"check_interval" default:"5m"`
		MinIndicesStore        string        `file:"min_indices_store" default:"10GB"`
		MinIndicesStorePercent float64       `file:"min_indices_store_percent" default:"10"`
		HighDiskUsagePercent   float64       `file:"high_disk_usage_percent" default:"85"`
		LowDiskUsagePercent    float64       `file:"low_disk_usage_percent" default:"70"`
		RolloverBodyFile       string        `file:"rollover_body_file"`
		minIndicesStore        int64
	} `file:"disk_clean"`

	LockKey string `file:"lock_key" default:"metric-index-task-lock"`
}

type provider struct {
	C   *config
	L   logs.Logger
	m   *IndexManager
	ctx servicehub.Context
}

var _ Index = (*IndexManager)(nil)

func (p *provider) Init(ctx servicehub.Context) error {
	p.ctx = ctx
	if int64(p.C.IndexCleanInterval) <= 0 {
		p.C.EnableIndexClean = false
	}
	if int64(p.C.RolloverInterval) <= 0 {
		p.C.EnableRollover = false
	}
	es := ctx.Service("elasticsearch").(elasticsearch.Interface)
	db := ctx.Service("mysql").(mysql.Interface).DB()
	p.m = NewIndexManager(p.C, es.Client(), es.URL(), db, p.L)

	if p.C.EnableIndexInit {
		err := p.initTemplate(es.Client())
		if err != nil {
			return err
		}
	}
	routes := ctx.Service("http-server", interceptors.CORS()).(httpserver.Router)
	err := p.intRoutes(routes)
	if err != nil {
		return fmt.Errorf("fail to init routes: %s", err)
	}
	return nil
}

func (p *provider) Start() error {
	mu, _ := p.ctx.Service("etcd-mutex").(mutex.Interface)
	var lock mutex.Mutex
	if mu != nil {
		lk, err := mu.New(context.Background(), p.C.LockKey)
		if err != nil {
			p.L.Error(err)
		}
		lock = lk
	}
	return p.m.Start(lock)
}

func (p *provider) Close() error { return p.m.Close() }

func (p *provider) Provide(ctx servicehub.DependencyContext, options ...interface{}) interface{} {
	return p.m
}

func init() {
	servicehub.Register("erda.core.monitor.metric.index", &servicehub.Spec{
		Services:     []string{"erda.core.monitor.metric.index"},
		Dependencies: []string{"elasticsearch", "mysql", "http-server"},
		Description:  "metrics index manager",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
