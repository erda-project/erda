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

// Package cmp Core components of multi-cloud management platform
package cmp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/version"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/conf"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/modules/cmp/endpoints"
	"github.com/erda-project/erda/modules/cmp/i18n"
	aliyun_resources "github.com/erda-project/erda/modules/cmp/impl/aliyun-resources"
	org_resource "github.com/erda-project/erda/modules/cmp/impl/org-resource"
	"github.com/erda-project/erda/modules/cmp/steve/middleware"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/dumpstack"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

func initialize(ctx context.Context) error {
	conf.Load()

	// set log formatter
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000000",
	})
	logrus.SetOutput(os.Stdout)

	// set the debug level of log
	debugLevel := logrus.InfoLevel
	if conf.Debug() {
		debugLevel = logrus.DebugLevel
	}
	logrus.SetLevel(debugLevel)

	dumpstack.Open()
	logrus.Infoln(version.String())

	server, err := do(ctx)
	if err != nil {
		return err
	}

	return server.ListenAndServe()
}

func do(ctx context.Context) (*httpserver.Server, error) {
	var redisCli *redis.Client

	db := dbclient.Open(dbengine.MustOpen())

	i18n.InitI18N()

	// cache etcd
	option := jsonstore.UseCacheEtcdStore(context.Background(), aliyun_resources.CloudResourcePrefix, 100)
	cachedJs, err := jsonstore.New(option)

	// etcd
	js, err := jsonstore.New()
	if err != nil {
		return nil, err
	}

	if conf.LocalMode() {
		redisCli = redis.NewClient(&redis.Options{
			Addr:     conf.RedisAddr(),
			Password: conf.RedisPwd(),
		})
	} else {
		redisCli = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    conf.RedisMasterName(),
			SentinelAddrs: strings.Split(conf.RedisSentinelAddrs(), ","),
			Password:      conf.RedisPwd(),
		})
	}
	if _, err := redisCli.Ping().Result(); err != nil {
		return nil, err
	}

	// init uc client
	uc := ucauth.NewUCClient(discover.UC(), conf.UCClientID(), conf.UCClientSecret())
	if conf.OryEnabled() {
		uc = ucauth.NewUCClient(conf.OryKratosPrivateAddr(), conf.OryCompatibleClientID(), conf.OryCompatibleClientSecret())
	}

	// init Bundle
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*60),
			)),
		bundle.WithPipeline(),
		bundle.WithScheduler(),
		bundle.WithMonitor(),
		bundle.WithCoreServices(),
		bundle.WithOrchestrator(),
		bundle.WithDiceHub(),
		bundle.WithEventBox(),
		bundle.WithClusterManager(),
	}
	bdl := bundle.New(bundleOpts...)

	o := org_resource.New(
		org_resource.WithDBClient(db),
		org_resource.WithUCClient(uc),
		org_resource.WithBundle(bdl),
		org_resource.WithRedisClient(redisCli),
	)

	ep, err := initEndpoints(ctx, db, js, cachedJs, bdl, o)
	if err != nil {
		return nil, err
	}

	if conf.EnableEss() {
		initServices(ep)
	}

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(append(ep.Routes()))

	authenticator := middleware.NewAuthenticator(bdl)
	auditor := middleware.NewAuditor(bdl)

	middlewares := middleware.Chain{
		authenticator.AuthMiddleware,
		auditor.AuditMiddleWare,
	}
	server.Router().PathPrefix("/api/k8s/clusters/{clusterName}").Handler(middlewares.Handler(ep.SteveAggregator))

	logrus.Infof("start the service and listen on address: %s", conf.ListenAddr())
	logrus.Info("starting cmp instance")

	// init cron job
	initCron(ep)

	return server, nil
}

func initEndpoints(ctx context.Context, db *dbclient.DBClient, js, cachedJS jsonstore.JsonStore, bdl *bundle.Bundle, o *org_resource.OrgResource) (*endpoints.Endpoints, error) {

	// compose endpoints
	ep := endpoints.New(
		ctx,
		db,
		js,
		cachedJS,
		endpoints.WithBundle(bdl),
		endpoints.WithOrgResource(o),
	)

	// Sync org resource task status
	go func() {
		ep.SyncTaskStatus(conf.TaskSyncDuration())
	}()

	// Clean job/deployment sync
	go func() {
		ep.TaskClean(conf.TaskCleanDuration())
	}()

	registerWebHook(bdl)

	return ep, nil
}

func initServices(ep *endpoints.Endpoints) {
	// run mns service, monitor mns messages & consume them
	ep.Mns.Run()
	ep.Ess.AutoScale()
}

// 初始化定时任务
func initCron(ep *endpoints.Endpoints) {
	// cron job to monitor pipeline created edge clusters
	go loop.New(loop.WithInterval(10 * time.Second)).Do(ep.GetCluster().MonitorCloudCluster)
}

func registerWebHook(bdl *bundle.Bundle) {
	// register pipeline tasks by webhook
	ev := apistructs.CreateHookRequest{
		Name:   "cmdb_pipeline_tasks",
		Events: []string{"pipeline_task", "pipeline_task_runtime"},
		URL:    strutil.Concat("http://", discover.CMP(), "/api/tasks"),
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}
	if err := bdl.CreateWebhook(ev); err != nil {
		logrus.Warnf("failed to register pipeline tasks event, (%v)", err)
	}

	clusterEv := apistructs.CreateHookRequest{
		Name:   "cmp-clusterhook",
		Events: []string{"cluster"},
		URL:    fmt.Sprintf("http://%s/api/clusterhook", discover.CMP()),
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}
	if err := bdl.CreateWebhook(clusterEv); err != nil {
		logrus.Warnf("failed to register cluster event, (%v)", err)
	}
}
