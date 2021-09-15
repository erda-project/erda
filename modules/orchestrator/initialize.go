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

// Package orchestrator 编排器
package orchestrator

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	infrahttpserver "github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/conf"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/endpoints"
	"github.com/erda-project/erda/modules/orchestrator/services/addon"
	"github.com/erda-project/erda/modules/orchestrator/services/deployment"
	"github.com/erda-project/erda/modules/orchestrator/services/domain"
	"github.com/erda-project/erda/modules/orchestrator/services/instance"
	"github.com/erda-project/erda/modules/orchestrator/services/migration"
	"github.com/erda-project/erda/modules/orchestrator/services/resource"
	"github.com/erda-project/erda/modules/orchestrator/services/runtime"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/goroutinepool"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/loop"
	// "terminus.io/dice/telemetry/promxp"
)

// Initialize 初始化应用启动服务.
func (p *provider) Initialize(ctx servicehub.Context) error {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	db := &dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: p.Orm,
		},
	}

	// init endpoints
	ep, err := p.initEndpoints(db)
	if err != nil {
		return err
	}

	bdl := bundle.New()
	// This server will never be started. Only the routes and locale loader are used by new http server
	server := httpserver.New(":0")
	server.WithLocaleLoader(bdl.GetLocaleLoader())
	// server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("orchestrator"))
	server.RegisterEndpoint(ep.Routes())

	ctx.Service("http-server").(infrahttpserver.Router).Any("/**", server.Router())

	// Limit only one instance of scheduler to do the cron jobs
	p.Election.OnLeader(func(ctx context.Context) {
		logrus.Infof("i'm the leader now")
		_ = initLeaderCron(ep, ctx)
		logrus.Infof("i resign the leader now")
	})

	// start cron jobs to sync addon & project infos
	go initCron(ep, ctx)

	return nil
}

// 初始化 Endpoints
func (p *provider) initEndpoints(db *dbclient.DBClient) (*endpoints.Endpoints, error) {
	// init pool
	pool := goroutinepool.New(conf.PoolSize())
	pool.Start()

	// init Bundle
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*60),
			)),
		bundle.WithCoreServices(),
		bundle.WithDiceHub(),
		bundle.WithEventBox(),
		bundle.WithScheduler(),
		bundle.WithCollector(),
		bundle.WithMonitor(),
		bundle.WithHepa(),
		bundle.WithCMP(),
		bundle.WithKMS(),
		bundle.WithOpenapi(),
		bundle.WithPipeline(),
		bundle.WithGittar(),
		bundle.WithMSP(),
		bundle.WithDOP(),
		bundle.WithClusterManager(),
	}
	bdl := bundle.New(bundleOpts...)

	encrypt := encryption.New(
		encryption.WithRSAScrypt(encryption.NewRSAScrypt(encryption.RSASecret{
			PublicKey:          conf.PublicKey(),
			PublicKeyDataType:  encryption.Base64,
			PrivateKey:         conf.PrivateKey(),
			PrivateKeyType:     encryption.PKCS1,
			PrivateKeyDataType: encryption.Base64,
		})))

	migration := migration.New(
		migration.WithBundle(bdl),
		migration.WithDBClient(db))

	resource := resource.New(
		resource.WithDBClient(db),
		resource.WithBundle(bdl),
	)

	// init addon service
	a := addon.New(
		addon.WithDBClient(db),
		addon.WithBundle(bdl),
		addon.WithResource(resource),
		addon.WithEnvEncrypt(encrypt),
		addon.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second, time.Second*60),
		)),
	)

	// init runtime service
	rt := runtime.New(
		runtime.WithDBClient(db),
		runtime.WithEventManager(p.EventManager),
		runtime.WithBundle(bdl),
		runtime.WithAddon(a))

	// init deployment service
	d := deployment.New(
		deployment.WithDBClient(db),
		deployment.WithEventManager(p.EventManager),
		deployment.WithBundle(bdl),
		deployment.WithAddon(a),
		deployment.WithMigration(migration),
		deployment.WithEncrypt(encrypt),
		deployment.WithResource(resource),
	)

	// init domain service
	dom := domain.New(
		domain.WithDBClient(db),
		domain.WithEventManager(p.EventManager),
		domain.WithBundle(bdl))

	ins := instance.New(
		instance.WithBundle(bdl),
	)

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithDBClient(db),
		endpoints.WithPool(pool),
		endpoints.WithQueue(p.PusherQueue),
		endpoints.WithEventManager(p.EventManager),
		endpoints.WithBundle(bdl),
		endpoints.WithRuntime(rt),
		endpoints.WithDeployment(d),
		endpoints.WithDomain(dom),
		endpoints.WithAddon(a),
		endpoints.WithInstance(ins),
		endpoints.WithEnvEncrypt(encrypt),
		endpoints.WithResource(resource),
		endpoints.WithMigration(migration),
	)

	return ep, nil
}

// 初始化 leader 定时任务, 单实例执行
func initLeaderCron(ep *endpoints.Endpoints, ctx context.Context) error {
	// cron for pushOn deployment
	go loop.New(loop.WithContext(ctx), loop.WithInterval(10*time.Second)).Do(ep.PushOnDeploymentPolling)
	go loop.New(loop.WithContext(ctx), loop.WithDeclineRatio(1.2), loop.WithInterval(50*time.Millisecond),
		loop.WithDeclineLimit(3*time.Second)).Do(ep.PushOnDeployment)
	go loop.New(loop.WithContext(ctx), loop.WithInterval(10*time.Second)).Do(ep.PushOnDeletingRuntimesPolling)
	go loop.New(loop.WithContext(ctx), loop.WithInterval(2*time.Second)).Do(ep.PushOnDeletingRuntimes)

	go loop.New(loop.WithContext(ctx), loop.WithInterval(10*time.Minute)).Do(ep.SyncAddonReferenceNum)

	go ep.CleanUnusedMigrationNs()
	go loop.New(loop.WithContext(ctx), loop.WithInterval(24*time.Hour)).Do(ep.CleanUnusedMigrationNs)

	go ep.SyncDeployAddon()

	go loop.New(loop.WithContext(ctx), loop.WithInterval(5*time.Minute)).Do(ep.SyncAddonResources)

	ep.FullGCLoop(ctx)

	return nil
}

type SharedCronjobRunner interface {
	SyncAddons() (bool, error)
	SyncProjects() (bool, error)
}

// 初始化定时任务
func initCron(ep SharedCronjobRunner, ctx context.Context) {
	go ep.SyncAddons()
	go loop.New(loop.WithContext(ctx), loop.WithInterval(10*time.Minute)).Do(ep.SyncAddons)
	go ep.SyncProjects()
	go loop.New(loop.WithContext(ctx), loop.WithInterval(5*time.Minute)).Do(ep.SyncProjects)
	<-ctx.Done()
}
