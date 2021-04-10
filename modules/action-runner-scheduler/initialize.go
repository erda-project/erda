//  Copyright (c) 2021 Terminus, Inc.
//
//  This program is free software: you can use, redistribute, and/or modify
//  it under the terms of the GNU Affero General Public License, version 3
//  or later ("AGPL"), as published by the Free Software Foundation.
//
//  This program is distributed in the hope that it will be useful, but WITHOUT
//  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
//  FITNESS FOR A PARTICULAR PURPOSE.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program. If not, see <http://www.gnu.org/licenses/>.

package action_runner_scheduler

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/action-runner-scheduler/conf"
	"github.com/erda-project/erda/modules/action-runner-scheduler/dbclient"
	"github.com/erda-project/erda/modules/action-runner-scheduler/endpoints"
	"github.com/erda-project/erda/modules/action-runner-scheduler/services/runnertask"
	"github.com/erda-project/erda/pkg/httpserver"
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

	bdl := bundle.New(bundle.WithCollector(), bundle.WithCMDB(), bundle.WithOpenapi())
	runnerTask := runnertask.New(runnertask.WithDBClient(db), runnertask.WithBundle(bdl))
	ep := endpoints.New(
		endpoints.WithRunnerTask(runnerTask),
		endpoints.WithBundle(bdl),
	)

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(ep.Routes())
	return server.ListenAndServe()
}
