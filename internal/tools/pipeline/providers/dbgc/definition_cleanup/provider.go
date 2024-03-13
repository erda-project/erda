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
	"fmt"
	"io"
	"os"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
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
	RetryTimes    int           `file:"retry_times" env:"PIPELINE_DEFINITION_CLEANUP_RETRY_TIMES" default:"3"`
	RetryInterval time.Duration `file:"retry_interval" env:"PIPELINE_DEFINITION_CLEANUP_RETRY_INTERVAL" default:"5s"`

	// ------- cleanup cronExpr ------
	CronExpr string `file:"cron_expr" env:"PIPELINE_DEFINITION_CLEANUP_CRON_EXPR" default:"0 15 2 * * ?"`

	// ------- dry run ---------
	DryRun         bool   `file:"dry_run" env:"PIPELINE_DEFINITION_CLEANUP_DRY_RUN" default:"false"`
	DryRunFilePath string `file:"file_path" env:"PIPELINE_DEFINITION_CLEANUP_DRY_RUN_FILEPATH" default:"/erda/dry-run"`
}

type provider struct {
	Cfg         *config
	Log         logs.Logger
	MySQL       mysqlxorm.Interface
	LW          leaderworker.Interface
	CronService cronpb.CronServiceServer `autowired:"erda.core.pipeline.cron.CronService" required:"true"`

	dbClient           *db.Client
	sourceDbClient     *sourcedb.Client
	definitionDbClient *definitiondb.Client
	cronDbClient       *crondb.Client
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.dbClient = &db.Client{Client: dbclient.Client{Engine: p.MySQL.DB()}}
	p.sourceDbClient = &sourcedb.Client{Interface: p.MySQL}
	p.definitionDbClient = &definitiondb.Client{Interface: p.MySQL}
	p.cronDbClient = &crondb.Client{Interface: p.MySQL}

	if p.Cfg.DryRun {
		// check if path exist
		if _, err := os.Stat(p.Cfg.DryRunFilePath); os.IsNotExist(err) {
			// dir is not exist
			os.Mkdir(p.Cfg.DryRunFilePath, 0755)
		}

		filepath := p.Cfg.DryRunFilePath + fmt.Sprintf("/dry-run_%s.log", time.Now().Format("2006_01_02"))
		// create file
		file, err := os.Create(filepath)
		if err != nil {
			p.Log.Error("create dry run log file err: %s", err)
		}

		os.Chmod(filepath, 0666)

		p.Log = logrusx.New()
		p.Log = p.Log.Sub(dryrunPrifix)
		p.Log.SetOutput(io.MultiWriter(os.Stdout, file))
		p.Log.Infof("Start Cleanup Definition in %s", p.Cfg.CronExpr)
	}

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
