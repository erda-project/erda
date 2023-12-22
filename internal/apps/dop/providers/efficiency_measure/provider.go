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

package efficiency_measure

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda/bundle"
	iterationdb "github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	issuedb "github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/reconciler/rutil"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/jsonstore"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type config struct {
	RefreshPerformanceBasicInfoDuration        time.Duration `file:"refresh_performance_basic_info_duration" env:"REFRESH_PERFORMANCE_BASIC_INFO_DURATION" default:"1h"`
	RefreshCheckPerformanceNumberFiledDuration time.Duration `file:"refresh_check_performance_number_filed_duration" env:"REFRESH_CHECK_PERFORMANCE_NUMBER_FILED_DURATION" default:"20m"`
	PerformanceMetricEtcdPrefixKey             string        `file:"performance_metric_etcd_prefix_key" env:"PERFORMANCE_METRIC_ETCD_PREFIX_KEY" default:"/devops/metrics/performance/"`
	OrgWhiteList                               []string      `file:"performance_measure_org_white_list" env:"PERFORMANCE_MEASURE_ORG_WHITE_LIST"`
	DemandStageList                            []string      `file:"demand_stage_list" env:"DEMAND_STAGE_LIST" default:"demandDesign,需求设计,架构设计,architectureDesign,需求调研"`
	ArchitectureStageList                      []string      `file:"architecture_stage_list" env:"ARCHITECTURE_STAGE_LIST" default:"交互设计,技术设计,UI设计"`
}

// +provider
type provider struct {
	Cfg         *config
	Log         logs.Logger
	bdl         *bundle.Bundle
	projDB      *dao.DBClient
	issueDB     *issuedb.DBClient
	iterationDB *iterationdb.DBClient
	Clickhouse  clickhouse.Interface `autowired:"clickhouse" optional:"true"`
	DB          *gorm.DB             `autowired:"mysql-gorm.v2-client"`

	Org      org.Interface
	IssueSvc query.Interface
	Election election.Interface `autowired:"etcd-election@efficiency-measure"`
	Js       jsonstore.JsonStore
	Register transport.Register

	errors                prometheus.Gauge
	propertySet           *propertyCache
	personalEfficiencySet *personalEfficiencyCache
}

type itemCollector struct {
	helper *provider
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithErdaServer())
	js, err := jsonstore.New()
	if err != nil {
		return fmt.Errorf("failed to init jsonstore, err: %v", err)
	}
	p.Js = js
	db, err := dao.NewDBClient()
	if err != nil {
		return err
	}
	p.projDB = db
	p.issueDB = &issuedb.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: db.DB,
		},
	}

	p.errors = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "efficiency_measure",
		Name:      "scrape_error",
		Help:      "1 if there was an error while getting efficiency measure metrics, 0 otherwise",
	})
	p.personalEfficiencySet = &personalEfficiencyCache{cache.New(cache.NoExpiration, cache.NoExpiration)}
	p.propertySet = &propertyCache{cache.New(cache.NoExpiration, cache.NoExpiration)}

	registry := prometheus.NewRegistry()
	registry.MustRegister(p)

	i := &itemCollector{
		helper: p,
	}
	registryIDs := prometheus.NewRegistry()
	registryIDs.MustRegister(i)

	p.Register.Add(http.MethodGet, "/personal-metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
	p.Register.Add(http.MethodGet, "/personal-metrics-item-ids", func(w http.ResponseWriter, r *http.Request) {
		promhttp.HandlerFor(registryIDs, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
	p.Register.Add(http.MethodPost, "/api/efficiency-measure/actions/query", p.queryPersonalEfficiency)
	p.Register.Add(http.MethodPost, "/api/personal-contribution/actions/query", p.queryPersonalContributors)
	p.Register.Add(http.MethodPost, "/api/func-points-trend/actions/query", p.queryFuncPointTrend)
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	p.Start(ctx)
	return nil
}

func (p *provider) Start(ctx context.Context) {
	go func() {
		rutil.ContinueWorking(ctx, p.Log, func(ctx context.Context) rutil.WaitDuration {
			if err := p.refreshBasicInfo(); err != nil {
				p.Log.Errorf("failed to refresh basic user projects, err: %v", err)
				return rutil.ContinueWorkingWithCustomInterval(time.Minute)
			}

			return rutil.ContinueWorkingWithDefaultInterval
		}, rutil.WithContinueWorkingDefaultRetryInterval(p.Cfg.RefreshPerformanceBasicInfoDuration))
	}()
	go func() {
		<-time.NewTimer(time.Duration(rand.Intn(5)) * time.Minute).C
		rutil.ContinueWorking(ctx, p.Log, func(ctx context.Context) rutil.WaitDuration {
			p.checkPersonalNumberFields()

			return rutil.ContinueWorkingWithDefaultInterval
		}, rutil.WithContinueWorkingDefaultRetryInterval(p.Cfg.RefreshCheckPerformanceNumberFiledDuration))
	}()
}

func init() {
	servicehub.Register("efficiency-measure", &servicehub.Spec{
		Services:    []string{"efficiency-measure"},
		Description: "efficiency measure",
		ConfigFunc:  func() interface{} { return &config{} },
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
