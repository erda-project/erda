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

package definition_cleanup

import (
	"context"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	crondb "github.com/erda-project/erda/internal/tools/pipeline/providers/cron/db"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/dbgc/db"
	definitiondb "github.com/erda-project/erda/internal/tools/pipeline/providers/definition/db"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/leaderworker"
	sourcedb "github.com/erda-project/erda/internal/tools/pipeline/providers/source/db"
)

type config struct {
	// ------ cleanup retry -----
	// default 5
	PipelineCleanupRetryTimes    int           `file:"pipeline_cleanup_retry_times" env:"PIPELINE_CLEANUP_RETRY_TIMES" default:"3"`
	PipelineCleanupRetryInterval time.Duration `file:"pipeline_cleanup_retry_interval" env:"PIPELINE_CLEANUP_RETRY_INTERVAL" default:"5s"`

	// ------- cleanup cronExpr ------
	PipelineCleanupCron string `file:"pipeline_cleanup_cron" env:"PIPELINE_CLEANUP_CRON" default:"0 15 2 * * *"`
}

type provider struct {
	Cfg                *config
	Log                logs.Logger
	dbClient           *db.Client
	sourceDbClient     *sourcedb.Client
	definitionDbClient *definitiondb.Client
	cronDbClient       *crondb.Client
	MySQL              mysqlxorm.Interface
	LW                 leaderworker.Interface
	CronService        cronpb.CronServiceServer `autowired:"erda.core.pipeline.cron.CronService" required:"true"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.dbClient = &db.Client{Client: dbclient.Client{Engine: p.MySQL.DB()}}
	p.sourceDbClient = &sourcedb.Client{Interface: p.MySQL}
	p.definitionDbClient = &definitiondb.Client{Interface: p.MySQL}
	p.cronDbClient = &crondb.Client{Interface: p.MySQL}

	return nil
}

func (p *provider) Run(ctx context.Context) error {
	p.LW.OnLeader(p.RepeatPipelineRecordCleanup)
	return nil
}

func init() {
	servicehub.Register("definition-cleanup", &servicehub.Spec{
		Services:     []string{"definition-cleanup"},
		Dependencies: nil,
		Description:  "pipeline definition/record/base/cron cleanup",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
