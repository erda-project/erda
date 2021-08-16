// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Package orchestrator 编排器
package orchestrator

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/conf"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/endpoints"
	"github.com/erda-project/erda/modules/orchestrator/events"
	"github.com/erda-project/erda/modules/orchestrator/queue"
	"github.com/erda-project/erda/modules/orchestrator/services/addon"
	"github.com/erda-project/erda/modules/orchestrator/services/deployment"
	"github.com/erda-project/erda/modules/orchestrator/services/domain"
	"github.com/erda-project/erda/modules/orchestrator/services/instance"
	"github.com/erda-project/erda/modules/orchestrator/services/migration"
	"github.com/erda-project/erda/modules/orchestrator/services/resource"
	"github.com/erda-project/erda/modules/orchestrator/services/runtime"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/goroutinepool"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/loop"
	// "terminus.io/dice/telemetry/promxp"
)

// Initialize 初始化应用启动服务.
func Initialize() error {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// init db
	db, err := dbclient.Open()
	defer db.Close()
	if err != nil {
		return err
	}

	// init endpoints
	ep, err := initEndpoints(db)
	if err != nil {
		return err
	}

	bdl := bundle.New()
	server := httpserver.New(conf.ListenAddr())
	server.WithLocaleLoader(bdl.GetLocaleLoader())
	// server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("orchestrator"))
	server.RegisterEndpoint(ep.Routes())

	if err := initCron(ep); err != nil {
		return err
	}

	logrus.Infof("start the service and listen on address: \"%s\"", conf.ListenAddr())
	logrus.Errorf("[alert] starting orchestrator instance")

	return server.ListenAndServe()
}

// 初始化 Endpoints
func initEndpoints(db *dbclient.DBClient) (*endpoints.Endpoints, error) {
	// init PusherQueue
	pq := queue.NewPusherQueue()

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

	// init EventManager
	evMgr := events.NewEventManager(1000, pq, db, bdl)
	evMgr.Start()

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
		runtime.WithEventManager(evMgr),
		runtime.WithBundle(bdl),
		runtime.WithAddon(a))

	// init deployment service
	d := deployment.New(
		deployment.WithDBClient(db),
		deployment.WithEventManager(evMgr),
		deployment.WithBundle(bdl),
		deployment.WithAddon(a),
		deployment.WithMigration(migration),
		deployment.WithEncrypt(encrypt),
		deployment.WithResource(resource),
	)

	// init domain service
	dom := domain.New(
		domain.WithDBClient(db),
		domain.WithEventManager(evMgr),
		domain.WithBundle(bdl))

	ins := instance.New(
		instance.WithBundle(bdl),
	)

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithDBClient(db),
		endpoints.WithPool(pool),
		endpoints.WithQueue(pq),
		endpoints.WithEventManager(evMgr),
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

// 初始化定时任务
func initCron(ep *endpoints.Endpoints) error {
	// cron for pushOn deployment
	go loop.New(loop.WithInterval(10 * time.Second)).Do(ep.PushOnDeploymentPolling)
	go loop.New(loop.WithDeclineRatio(1.2), loop.WithInterval(50*time.Millisecond), loop.WithDeclineLimit(3*time.Second)).
		Do(ep.PushOnDeployment)
	go loop.New(loop.WithInterval(10 * time.Second)).Do(ep.PushOnDeletingRuntimesPolling)
	go loop.New(loop.WithInterval(2 * time.Second)).Do(ep.PushOnDeletingRuntimes)

	go ep.SyncAddons()
	go loop.New(loop.WithInterval(10 * time.Minute)).Do(ep.SyncAddons)

	//go loop.New(loop.WithInterval(10 * time.Minute)).Do(ep.RemoveAddons)

	go ep.SyncProjects()
	go loop.New(loop.WithInterval(5 * time.Minute)).Do(ep.SyncProjects)

	go loop.New(loop.WithInterval(10 * time.Minute)).Do(ep.SyncAddonReferenceNum)

	go ep.CleanUnusedMigrationNs()
	go loop.New(loop.WithInterval(24 * time.Hour)).Do(ep.CleanUnusedMigrationNs)

	go ep.SyncDeployAddon()

	go loop.New(loop.WithInterval(5 * time.Minute)).Do(ep.SyncAddonResources)

	ep.FullGCLoop()

	return nil
}
