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

package cleaner

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	cfgpkg "github.com/recallsong/go-utils/config"
	"github.com/recallsong/go-utils/lang/size"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

type (
	// RetentionStrategy .
	RetentionStrategy interface {
		GetTTL(*loader.IndexEntry) time.Duration
	}
	// RetentionStrategyLoader .
	RetentionStrategyLoader interface {
		Loading(ctx context.Context)
	}
	// Interface .
	Interface interface {
		CleanIndices(ctx context.Context, filter loader.Matcher) error
	}
)

type (
	config struct {
		RequestTimeout time.Duration `file:"request_timeout" default:"1m"`
		CheckInterval  time.Duration `file:"check_interval" default:"1h"`
		PrintOnly      bool          `file:"print_onluy"`
		DiskClean      struct {
			Enable                 bool          `file:"enable"`
			CheckInterval          time.Duration `file:"check_interval" default:"5m"`
			MinIndicesStore        string        `file:"min_indices_store" default:"10GB"`
			MinIndicesStorePercent float64       `file:"min_indices_store_percent" default:"10"`
			HighDiskUsagePercent   float64       `file:"high_disk_usage_percent" default:"85"`
			LowDiskUsagePercent    float64       `file:"low_disk_usage_percent" default:"70"`
			RolloverBodyFile       string        `file:"rollover_body_file"`
			RolloverAliasSuffix    string        `file:"rollover_alias_suffix"`
		} `file:"disk_clean"`
	}
	provider struct {
		Cfg        *config
		Log        logs.Logger
		election   election.Interface
		loader     loader.Interface
		retentions RetentionStrategy

		clearCh chan *clearRequest

		// for disk clean
		minIndicesStoreInDisk    int64
		rolloverBodyForDiskClean string
	}
)

var _ Interface = (*provider)(nil)

func (p *provider) Init(ctx servicehub.Context) error {
	loader, err := loader.Find(ctx, p.Log, true)
	if err != nil {
		return err
	}
	p.loader = loader

	election, err := index.FindElection(ctx, p.Log, true)
	if err != nil {
		return err
	}
	p.election = election

	if err := p.initRetentionStrategy(ctx); err != nil {
		return err
	}
	if loader, ok := p.retentions.(RetentionStrategyLoader); ok {
		p.election.OnLeader(loader.Loading)
	}

	if int64(p.Cfg.CheckInterval) <= 0 {
		return fmt.Errorf("invalid check_interval: %v", p.Cfg.CheckInterval)
	}
	if !p.loader.QueryIndexTimeRange() {
		p.Log.Warnf("index clean is enable, but QueryIndexTimeRange of elasticsearch.index.loader is disable")
	}
	p.election.OnLeader(p.runCleanIndices)

	if p.Cfg.DiskClean.Enable {
		if int64(p.Cfg.DiskClean.CheckInterval) <= 0 {
			return fmt.Errorf("invalid disk_clean.check_interval: %v", p.Cfg.DiskClean.CheckInterval)
		}

		// rollover body for disk clean
		if len(p.Cfg.DiskClean.RolloverBodyFile) > 0 {
			body, err := ioutil.ReadFile(p.Cfg.DiskClean.RolloverBodyFile)
			if err != nil {
				return fmt.Errorf("failed to load rollover body file for disk clean: %s", err)
			}
			body = cfgpkg.EscapeEnv(body)
			p.rolloverBodyForDiskClean = string(body)
			if len(p.rolloverBodyForDiskClean) <= 0 {
				return fmt.Errorf("RolloverBody is empty for disk clean")
			}
			var m map[string]interface{}
			err = json.NewDecoder(strings.NewReader(p.rolloverBodyForDiskClean)).Decode(&m)
			if err != nil {
				return fmt.Errorf("invalid RolloverBody for disk clean: %v", string(body))
			}
			p.Log.Info("load rollover body for disk clean: \n", p.rolloverBodyForDiskClean)
		}

		minIndicesStore, err := size.ParseBytes(p.Cfg.DiskClean.MinIndicesStore)
		if err != nil {
			return fmt.Errorf("invalid min_indices_store: %s", err)
		}
		p.minIndicesStoreInDisk = minIndicesStore

		// run disk clean task on leader node
		p.election.OnLeader(p.runDiskCheckAndClean)
	}

	// init manager routes
	routeRrefix := "/api/elasticsearch/index"
	if len(ctx.Label()) > 0 {
		routeRrefix = routeRrefix + "/" + ctx.Label()
	} else {
		routeRrefix = routeRrefix + "/-"
	}
	routes := ctx.Service("http-router", interceptors.CORS()).(httpserver.Router)
	err = p.intRoutes(routes, routeRrefix)
	if err != nil {
		return fmt.Errorf("failed to init routes: %s", err)
	}
	return nil
}

func (p *provider) initRetentionStrategy(ctx servicehub.Context) error {
	obj, name := index.FindService(ctx, "elasticsearch.index.retention-strategy")
	if obj == nil {
		return fmt.Errorf("%q is required", name)
	}
	rs, ok := obj.(RetentionStrategy)
	if !ok {
		return fmt.Errorf("%q is not RetentionStrategy", name)
	}
	p.retentions = rs
	p.Log.Debugf("use RetentionStrategy(%q) for index clean", name)
	return nil
}

func init() {
	servicehub.Register("elasticsearch.index.cleaner", &servicehub.Spec{
		Services:     []string{"elasticsearch.index.cleaner"},
		Dependencies: []string{"http-router", "elasticsearch.index.loader", "elasticsearch.index.retention-strategy", "etcd-election"},
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{
				clearCh: make(chan *clearRequest),
			}
		},
	})
}
