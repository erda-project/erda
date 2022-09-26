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
	"fmt"
	"os"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/pipeline/providers/reconciler"

	"github.com/erda-project/erda/internal/tools/pipeline/providers/cron/compensator"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/websocket"
	"github.com/erda-project/erda/internal/tools/pipeline/aop"
	"github.com/erda-project/erda/internal/tools/pipeline/conf"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/endpoints"
	"github.com/erda-project/erda/internal/tools/pipeline/events"
	"github.com/erda-project/erda/internal/tools/pipeline/metrics"
	"github.com/erda-project/erda/internal/tools/pipeline/pexpr/pexpr_params"
	"github.com/erda-project/erda/internal/tools/pipeline/pipengine/pvolumes"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/pipelinefunc"
	"github.com/erda-project/erda/internal/tools/pipeline/services/pipelinesvc"
	"github.com/erda-project/erda/pkg/dumpstack"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
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
	queryStringDecoder.ZeroEmpty(true)

	// websocket publisher
	publisher, err := websocket.NewPublisher()
	if err != nil {
		return err
	}

	// bundle
	bdl := bundle.New(bundle.WithAllAvailableClients())

	// init services
	pipelineSvc := pipelinesvc.New(dbClient, bdl, p.ClusterInfo)

	// init CallbackActionFunc
	pipelinefunc.CallbackActionFunc = p.PipelineSvc.DealPipelineCallbackOfAction

	// todo resolve cycle import here through better module architecture
	pipelineFuncs := reconciler.PipelineSvcFuncs{
		MergePipelineYmlTasks:                   p.PipelineSvc.MergePipelineYmlTasks,
		HandleQueryPipelineYamlBySnippetConfigs: p.PipelineSvc.HandleQueryPipelineYamlBySnippetConfigs,
		MakeSnippetPipeline4Create:              p.PipelineSvc.MakeSnippetPipeline4Create,
		CreatePipelineGraph:                     p.PipelineSvc.CreatePipelineGraph,
		PreCheck:                                p.PipelineSvc.PreCheck,
		ConvertSnippetConfig2String:             p.PipelineSvc.ConvertSnippetConfig2String,
	}
	p.Reconciler.InjectLegacyFields(&pipelineFuncs)

	if err := registerSnippetClient(dbClient); err != nil {
		return err
	}

	pvolumes.Initialize(dbClient)
	pexpr_params.Initialize(dbClient)

	// init endpoints
	ep := endpoints.New(
		endpoints.WithDBClient(dbClient),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithCrondSvc(p.CronDaemon),
		endpoints.WithPipelineSvc(pipelineSvc),
		endpoints.WithClusterInfo(p.ClusterInfo),
		endpoints.WithMysql(p.MySQL),
	)

	p.CronDaemon.WithPipelineFunc(p.PipelineSvc.CreateV2)
	p.CronCompensate.WithPipelineFunc(compensator.PipelineFunc{CreatePipeline: p.PipelineSvc.CreateV2, RunPipeline: p.PipelineRun.RunOnePipeline})

	//server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("pipeline"))
	server := httpserver.New("")
	server.RegisterEndpoint(ep.Routes())
	if err := server.RegisterToNewHttpServerRouter(p.Router); err != nil {
		return err
	}

	// 加载 event manager
	events.Initialize(bdl, publisher, dbClient, p.EdgeRegister, p.Org)

	// aop
	aop.Initialize(bdl, dbClient, p.ReportSvc)

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
