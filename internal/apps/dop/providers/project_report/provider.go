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

package project_report

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
	RefreshBasicInfoDuration        time.Duration `file:"refresh_basic_info_duration" env:"REFRESH_BASIC_INFO_DURATION" default:"1h"`
	RefreshCheckNumberFiledDuration time.Duration `file:"refresh_check_number_filed_duration" env:"REFRESH_CHECK_NUMBER_FILED_DURATION" default:"20m"`
	IterationMetricEtcdPrefixKey    string        `file:"iteration_metric_etcd_prefix_key" env:"ITERATION_METRIC_ETCD_PREFIX_KEY" default:"/devops/metrics/iteration/"`
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

	Org      org.Interface
	IssueSvc query.Interface
	Election election.Interface `autowired:"etcd-election@project-management-report"`
	Js       jsonstore.JsonStore
	Register transport.Register
	DB       *gorm.DB `autowired:"mysql-gorm.v2-client"`

	orgSet       *orgCache
	projectSet   *projectCache
	iterationSet *iterationCache
}

type iterationCollector struct {
	helper *PrometheusCollector
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
	p.iterationDB = &iterationdb.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: db.DB,
		},
	}

	pr := &PrometheusCollector{
		infoProvider:        p,
		iterationLabelsFunc: p.iterationLabelsFunc,
		errors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "iteration",
			Name:      "scrape_error",
			Help:      "1 if there was an error while getting iteration metrics, 0 otherwise",
		}),
		iterationMetrics: allIterationMetrics,
	}
	prometheus.MustRegister(pr)
	i := &iterationCollector{
		helper: pr,
	}
	registryIDs := prometheus.NewRegistry()
	registryIDs.MustRegister(i)

	p.orgSet = &orgCache{cache.New(cache.NoExpiration, cache.NoExpiration)}
	p.projectSet = &projectCache{cache.New(cache.NoExpiration, cache.NoExpiration)}
	p.iterationSet = &iterationCache{cache.New(cache.NoExpiration, cache.NoExpiration)}

	p.Register.Add(http.MethodGet, "/metrics-item-ids", func(w http.ResponseWriter, r *http.Request) {
		promhttp.HandlerFor(registryIDs, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
	p.Register.Add(http.MethodPost, "/api/project-report/actions/query", p.queryProjectReport)
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	p.Start(ctx)
	return nil
}

func (p *provider) Start(ctx context.Context) {
	go func() {
		rutil.ContinueWorking(ctx, p.Log, func(ctx context.Context) rutil.WaitDuration {
			if err := p.refreshBasicIterations(); err != nil {
				p.Log.Errorf("failed to refresh basic iterations, err: %v", err)
				return rutil.ContinueWorkingWithCustomInterval(time.Minute)
			}

			return rutil.ContinueWorkingWithDefaultInterval
		}, rutil.WithContinueWorkingDefaultRetryInterval(p.Cfg.RefreshBasicInfoDuration))
	}()
	go func() {
		<-time.NewTimer(time.Duration(rand.Intn(10)) * time.Minute).C
		rutil.ContinueWorking(ctx, p.Log, func(ctx context.Context) rutil.WaitDuration {
			p.checkIterationNumberFields()

			return rutil.ContinueWorkingWithDefaultInterval
		}, rutil.WithContinueWorkingDefaultRetryInterval(p.Cfg.RefreshCheckNumberFiledDuration))
	}()
}

func (p *provider) GetRequestedIterationsInfo() (map[uint64]*IterationInfo, error) {
	result := make(map[uint64]*IterationInfo)
	p.iterationSet.Iterate(func(key string, value interface{}) error {
		iterationInfo := value.(*IterationInfo)
		if iterationInfo.IterationMetricFields == nil || iterationInfo.IterationMetricFields.IsValid() {
			fields, _ := p.getIterationFields(iterationInfo)
			iterationInfo.IterationMetricFields = fields
		}
		result[iterationInfo.Iteration.ID] = iterationInfo
		return nil
	})
	return result, nil
}

func init() {
	servicehub.Register("project-management-report", &servicehub.Spec{
		Services:    []string{"project-management-report"},
		Description: "project delivery management report",
		ConfigFunc:  func() interface{} { return &config{} },
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
