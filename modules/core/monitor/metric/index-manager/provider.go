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
	"io/ioutil"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
	cfgpkg "github.com/recallsong/go-utils/config"
	"github.com/recallsong/go-utils/lang/size"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	indexloader "github.com/erda-project/erda/modules/core/monitor/metric/index-loader"
	"github.com/erda-project/erda/pkg/router"
)

type config struct {
	RequestTimeout time.Duration `file:"request_timeout" default:"2m" env:"METRIC_REQUEST_TIMEOUT"`

	DefaultNamespace string `file:"default_namespace" env:"METRIC_DEFAULT_NAMESPACE"`
	Namespaces       []struct {
		Name      string             `file:"name"`
		Tags      []*router.KeyValue `file:"tags"`
		Namespace string             `file:"namespace"`
	} `file:"namespaces"`

	IndexType string `file:"index_type" env:"METRIC_INDEX_TYPE"`

	EnableIndexInit   bool   `file:"enable_index_init"`
	IndexTemplateName string `file:"index_template_name" default:"spot_metric_template"`
	IndexTemplatePath string `file:"index_template_path"`

	EnableRollover   bool          `file:"enable_rollover"`
	RolloverBodyFile string        `file:"rollover_body_file"`
	RolloverInterval time.Duration `file:"rollover_interval"`

	LoadIndexTTLFollowClean  bool          `file:"load_index_ttl_follow_clean"`
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
	} `file:"disk_clean"`
}

type provider struct {
	Cfg      *config
	Log      logs.Logger
	Loader   indexloader.Interface   `autowired:"erda.core.monitor.metric.index-loader"`
	ES       elasticsearch.Interface `autowired:"elasticsearch"`
	DB       *gorm.DB                `autowired:"mysql-client"`
	Election election.Interface      `autowired:"etcd-election@index" optional:"true"`

	clearCh               chan *clearRequest
	minIndicesStoreInDisk int64

	rolloverBody             string
	rolloverBodyForDiskClean string

	created     map[string]bool
	createdLock sync.Mutex

	iconfig          atomic.Value // index setting
	namespaces       *router.Router
	defaultNamespace string
	indexPrefix      string
}

var _ Interface = (*provider)(nil)

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Election == nil && (p.Cfg.EnableIndexClean || p.Cfg.EnableRollover) {
		return fmt.Errorf("etcd-election is required")
	}
	p.Loader.WatchLoadEvent(p.onIndicesReloaded)
	p.indexPrefix = p.Loader.IndexPrefix()
	p.namespaces = newNamespaceRouter(p.Cfg)
	p.defaultNamespace = normalizeIndexSegmentName(p.Cfg.DefaultNamespace)

	// index rollover
	if int64(p.Cfg.RolloverInterval) <= 0 {
		p.Cfg.EnableRollover = false
	}
	if p.Cfg.EnableRollover {
		body, err := ioutil.ReadFile(p.Cfg.RolloverBodyFile)
		if err != nil {
			return fmt.Errorf("fail to load rollover body file: %s", err)
		}
		body = cfgpkg.EscapeEnv(body)
		p.rolloverBody = string(body)
		if len(p.rolloverBody) <= 0 {
			return fmt.Errorf("invalid RolloverBody")
		}
		p.Log.Info("load rollover body: \n", p.rolloverBody)
		p.Election.OnLeader(p.runIndexRollover)
	}

	if int64(p.Cfg.IndexCleanInterval) <= 0 {
		p.Cfg.EnableIndexClean = false
	}
	if p.Cfg.EnableIndexClean {
		if !p.Cfg.EnableRollover && !p.Loader.QueryIndexTimeRange() {
			p.Log.Warnf("index clean is enable, but QueryIndexTimeRange of index-loader is disable")
		}
		// index clean
		if int64(p.Cfg.IndexCleanInterval) <= 0 {
			return fmt.Errorf("invalid IndexCleanInterval: %v", p.Cfg.IndexCleanInterval)
		}
		p.Election.OnLeader(p.runCleanIndices)

		// index clean by disk check
		if p.Cfg.DiskClean.CheckInterval <= 0 {
			p.Cfg.DiskClean.EnableIndexCleanByDisk = false
		}
		if p.Cfg.DiskClean.EnableIndexCleanByDisk {
			if p.Cfg.EnableRollover {
				body, err := ioutil.ReadFile(p.Cfg.DiskClean.RolloverBodyFile)
				if err != nil {
					return fmt.Errorf("failed to load rollover body file for disk clean: %s", err)
				}
				body = cfgpkg.EscapeEnv(body)
				p.rolloverBodyForDiskClean = string(body)
				if len(p.rolloverBodyForDiskClean) <= 0 {
					return fmt.Errorf("invalid RolloverBody for disk clean")
				}
				p.Log.Info("load rollover body for disk clean: \n", p.rolloverBodyForDiskClean)
			}
			minIndicesStore, err := size.ParseBytes(p.Cfg.DiskClean.MinIndicesStore)
			if err != nil {
				return fmt.Errorf("invalid min_indices_store: %s", err)
			}
			p.minIndicesStoreInDisk = minIndicesStore
			p.Election.OnLeader(p.runDiskCheckAndClean)
		}
	}

	// load index ttl
	if p.Cfg.LoadIndexTTLFromDatabase {
		if int64(p.Cfg.TTLReloadInterval) <= 0 {
			return fmt.Errorf("invalid TTLReloadInterval: %v", p.Cfg.TTLReloadInterval)
		}
		if p.Cfg.EnableIndexClean && p.Cfg.LoadIndexTTLFollowClean {
			p.Election.OnLeader(func(ctx context.Context) {
				p.runLoadTTL(ctx)
			})
		} else {
			ctx.AddTask(p.runLoadTTL, servicehub.WithTaskName("indices ttl config loader"))
		}
	}

	if p.Cfg.EnableIndexInit {
		err := p.initTemplate(p.ES.Client())
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

func (p *provider) onIndicesReloaded(indices map[string]*indexloader.IndexGroup) {
	p.createdLock.Lock()
	if len(p.created) > 0 {
		p.created = make(map[string]bool)
	}
	p.createdLock.Unlock()
}

func init() {
	servicehub.Register("erda.core.monitor.metric.index-manager", &servicehub.Spec{
		Services:     []string{"erda.core.monitor.metric.index-manager"},
		Dependencies: []string{"http-router"},
		Description:  "metrics index manager",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{
				created: make(map[string]bool),
				clearCh: make(chan *clearRequest),
			}
		},
	})
}
