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

// Package pipeline 流水线
package pipeline

import (
	"context"
	"fmt"
	"os"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/endpoints"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/metrics"
	"github.com/erda-project/erda/modules/pipeline/pexpr/pexpr_params"
	"github.com/erda-project/erda/modules/pipeline/pipengine/pvolumes"
	"github.com/erda-project/erda/modules/pipeline/pkg/pipelinefunc"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler/legacy/reconciler"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/appsvc"
	"github.com/erda-project/erda/modules/pipeline/services/buildartifactsvc"
	"github.com/erda-project/erda/modules/pipeline/services/buildcachesvc"
	"github.com/erda-project/erda/modules/pipeline/services/crondsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/services/permissionsvc"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinecronsvc"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinesvc"
	"github.com/erda-project/erda/modules/pipeline/services/queuemanage"
	"github.com/erda-project/erda/modules/pipeline/services/reportsvc"
	"github.com/erda-project/erda/modules/pkg/websocket"
	"github.com/erda-project/erda/pkg/dumpstack"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/pipeline_network_hook_client"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
	// "terminus.io/dice/telemetry/promxp"
)

func (p *provider) Init(ctx servicehub.Context) error {
	return p.Initialize()
}

// Initialize 初始化应用启动服务.
func (p *provider) Initialize() error {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000000",
	})
	logrus.SetOutput(os.Stdout)

	dumpstack.Open()
	logrus.Infoln(version.String())

	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("DEBUG MODE")
	}

	err := p.do()
	if err != nil {
		return err
	}
	return nil
}

func (p *provider) do() error {

	// metrics
	metrics.Initialize(p.MetricReport)

	// db client
	dbClient, err := dbclient.New()
	if err != nil {
		return err
	}

	// queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	// websocket publisher
	publisher, err := websocket.NewPublisher()
	if err != nil {
		return err
	}

	// etcd
	js, err := jsonstore.New()
	if err != nil {
		return err
	}
	etcdctl, err := etcd.New()
	if err != nil {
		return err
	}

	// bundle
	bdl := bundle.New(bundle.WithAllAvailableClients())

	// init services
	appSvc := appsvc.New(bdl)
	buildArtifactSvc := buildartifactsvc.New(dbClient)
	buildCacheSvc := buildcachesvc.New(dbClient)
	permissionSvc := permissionsvc.New(bdl)
	crondSvc := crondsvc.New(dbClient, bdl, js)
	actionAgentSvc := actionagentsvc.New(dbClient, bdl, js, etcdctl)
	extMarketSvc := extmarketsvc.New(bdl)
	pipelineCronSvc := pipelinecronsvc.New(dbClient, crondSvc)
	reportSvc := reportsvc.New(reportsvc.WithDBClient(dbClient))
	queueManage := queuemanage.New(queuemanage.WithDBClient(dbClient))

	// init services
	pipelineSvc := pipelinesvc.New(appSvc, crondSvc, actionAgentSvc, extMarketSvc, pipelineCronSvc,
		permissionSvc, queueManage, dbClient, bdl, publisher, p.Engine, js, etcdctl)
	pipelineSvc.WithCmsService(p.CmsService)

	// todo resolve cycle import here through better module architecture
	pipelineFun := &reconciler.PipelineSvcFunc{
		CronNotExecuteCompensate:                pipelineSvc.CronNotExecuteCompensateById,
		MergePipelineYmlTasks:                   pipelineSvc.MergePipelineYmlTasks,
		HandleQueryPipelineYamlBySnippetConfigs: pipelineSvc.HandleQueryPipelineYamlBySnippetConfigs,
		MakeSnippetPipeline4Create:              pipelineSvc.MakeSnippetPipeline4Create,
		CreatePipelineGraph:                     pipelineSvc.CreatePipelineGraph,
	}
	// init CallbackActionFunc
	pipelinefunc.CallbackActionFunc = pipelineSvc.DealPipelineCallbackOfAction

	r, err := reconciler.New(js, etcdctl, bdl, dbClient, actionAgentSvc, extMarketSvc, pipelineFun)
	if err != nil {
		return fmt.Errorf("failed to init reconciler, err: %v", err)
	}
	p.Reconciler.InjectLegacyReconciler(r)

	if err := registerSnippetClient(dbClient); err != nil {
		return err
	}
	if err := pipeline_network_hook_client.RegisterLifecycleHookClient(dbClient); err != nil {
		return err
	}

	pvolumes.Initialize(dbClient)
	pexpr_params.Initialize(dbClient)

	// init endpoints
	ep := endpoints.New(
		endpoints.WithDBClient(dbClient),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithAppSvc(appSvc),
		endpoints.WithBuildArtifactSvc(buildArtifactSvc),
		endpoints.WithBuildCacheSvc(buildCacheSvc),
		endpoints.WithPermissionSvc(permissionSvc),
		endpoints.WithCrondSvc(crondSvc),
		endpoints.WithActionAgentSvc(actionAgentSvc),
		endpoints.WithExtMarketSvc(extMarketSvc),
		endpoints.WithPipelineCronSvc(pipelineCronSvc),
		endpoints.WithPipelineSvc(pipelineSvc),
		endpoints.WithReportSvc(reportSvc),
		endpoints.WithQueueManage(queueManage),
		endpoints.WithQueueManager(p.QueueManager),
		endpoints.WithEngine(p.Engine),
	)

	//server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("pipeline"))
	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(ep.Routes())
	p.Router.Any("/**", server.Router())

	// 加载 event manager
	events.Initialize(bdl, publisher, dbClient)

	// aop
	aop.Initialize(bdl, dbClient, reportSvc)

	// TODO refactor cron daemon
	p.LeaderWorker.OnLeader(func(ctx context.Context) {
		pipelineSvc.DoCrondAbout(ctx)
	})

	return nil
}

func registerSnippetClient(dbclient *dbclient.Client) error {

	list, err := dbclient.FindSnippetClientList()
	if err != nil {
		return fmt.Errorf("not find snippet client list: error %v", err)
	}

	clientMap := make(map[string]*apistructs.DicePipelineSnippetClient)
	for _, v := range list {
		clientMap[v.Name] = &apistructs.DicePipelineSnippetClient{
			Name: v.Name,
			ID:   v.ID,
			Host: v.Host,
			Extra: apistructs.PipelineSnippetClientExtra{
				UrlPathPrefix: v.Extra.UrlPathPrefix,
			},
		}
	}
	pipeline_snippet_client.SetSnippetClientMap(clientMap)
	return nil
}
