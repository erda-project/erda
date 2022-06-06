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

package action_runner_scheduler

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/pipeline/action-runner-scheduler/conf"
	"github.com/erda-project/erda/internal/tools/pipeline/action-runner-scheduler/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/action-runner-scheduler/endpoints"
	"github.com/erda-project/erda/internal/tools/pipeline/action-runner-scheduler/services/runnertask"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// Initialize initialize module.
func Initialize() error {
	conf.Load()

	// init db
	db, err := dbclient.Open()
	defer db.Close()
	if err != nil {
		return err
	}

	bdl := bundle.New(bundle.WithCollector(), bundle.WithCoreServices(), bundle.WithOpenapi())
	runnerTask := runnertask.New(runnertask.WithDBClient(db), runnertask.WithBundle(bdl))
	ep := endpoints.New(
		endpoints.WithRunnerTask(runnerTask),
		endpoints.WithBundle(bdl),
	)

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(ep.Routes())
	return server.ListenAndServe()
}
