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

package dop

import (
	"context"
	"embed"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/base/version"
	componentprotocol "github.com/erda-project/erda-infra/providers/component-protocol"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/i18n"
	cmspb "github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/pipeline/providers/definition_client"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/dumpstack"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

//go:embed component-protocol/scenarios
var scenarioFS embed.FS

type provider struct {
	Log logs.Logger

	PipelineCms cmspb.CmsServiceServer      `autowired:"erda.core.pipeline.cms.CmsService" optional:"true"`
	PipelineDs  definition_client.Processor `autowired:"erda.core.pipeline.definition-process-client"`

	Protocol componentprotocol.Interface
	Tran     i18n.Translator `translator:"component-protocol"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Log.Info("init dop")

	// component-protocol
	p.Log.Info("init component-protocol")
	p.Protocol.SetI18nTran(p.Tran) // use custom i18n translator
	// compatible for legacy protocol context bundle
	bdl.Init(
		// bundle.WithDOP(), // TODO change to internal method invoke in component-protocol
		bundle.WithHepa(),
		bundle.WithOrchestrator(),
		bundle.WithEventBox(),
		bundle.WithGittar(),
		bundle.WithPipeline(),
		bundle.WithMonitor(),
		bundle.WithCollector(),
		bundle.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second*15, time.Duration(conf.BundleTimeoutSecond())*time.Second), // bundle 默认 (time.Second, time.Second*3)
		)),
		bundle.WithKMS(),
		bundle.WithCoreServices(),
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*90),
				httpclient.WithEnableAutoRetry(false),
			)),
		// TODO remove it after internal bundle invoke inside cp issue-manage adjusted
		bundle.WithCustom(discover.EnvDOP, "localhost:9527"),
	)
	p.Protocol.WithContextValue(types.GlobalCtxKeyBundle, bdl.Bdl)
	protocol.MustRegisterProtocolsFromFS(scenarioFS)
	p.Log.Info("init component-protocol done")

	return nil
}

func (p *provider) Run(ctx context.Context) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000000",
	})
	logrus.SetOutput(os.Stdout)

	dumpstack.Open()
	logrus.Infoln(version.String())

	return p.Initialize()
}

func init() {
	servicehub.Register("dop", &servicehub.Spec{
		Services: []string{"dop"},
		Creator:  func() servicehub.Provider { return &provider{} },
	})
}
