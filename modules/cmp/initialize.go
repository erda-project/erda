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
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/autoscanner"
	"github.com/erda-project/erda/modules/cmp/conf"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/modules/cmp/endpoints"
	"github.com/erda-project/erda/modules/cmp/endpoints/kubernetes"
	"github.com/erda-project/erda/modules/cmp/i18n"
	aliyun_resources "github.com/erda-project/erda/modules/cmp/impl/aliyun-resources"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/dumpstack"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/strutil"
)

func initialize() error {
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

	server, err := do()
	if err != nil {
		return nil
	}

	return server.ListenAndServe()
}

func do() (*httpserver.Server, error) {
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
	}
	bdl := bundle.New(bundleOpts...)

	ep, err := initEndpoints(db, js, cachedJs, bdl)
	if err != nil {
		return nil, err
	}
	if conf.EnableEss() {
		initServices(ep)
	}

	k8sep := newKubernetesEndpoints(bdl)

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(append(ep.Routes(), k8sep.Routers()...))

	logrus.Infof("start the service and listen on address: %s", conf.ListenAddr())
	logrus.Info("starting cmp instance")

	// autoScanner will scan expired cmp time
	as := autoscanner.New(db, bdl)
	logrus.Info("start autoScanner to scan expired cmp cluster")
	go as.Run()

	return server, nil
}

func initEndpoints(db *dbclient.DBClient, js, cachedJS jsonstore.JsonStore, bdl *bundle.Bundle) (*endpoints.Endpoints, error) {

	// compose endpoints
	ep := endpoints.New(
		db,
		js,
		cachedJS,
		endpoints.WithBundle(bdl),
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

func newKubernetesEndpoints(bdl *bundle.Bundle) *kubernetes.Endpoints {
	return kubernetes.New(bdl)
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
}
