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

package main

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/conf"
	"github.com/erda-project/erda/pkg/common"

	_ "github.com/erda-project/erda/modules/core/monitor/alert/storage/alert-record"
	// modules
	_ "github.com/erda-project/erda/modules/core/monitor/log/storage"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/storage"
	_ "github.com/erda-project/erda/modules/monitor/notify/storage/notify-record"
	_ "github.com/erda-project/erda/modules/msp/apm/browser"
	_ "github.com/erda-project/erda/modules/msp/apm/trace/storage"

	// providers
	_ "github.com/erda-project/erda-infra/providers/cassandra"
	_ "github.com/erda-project/erda-infra/providers/elasticsearch"
	_ "github.com/erda-project/erda-infra/providers/health"
	_ "github.com/erda-project/erda-infra/providers/kafka"
	_ "github.com/erda-project/erda-infra/providers/mysql"
	_ "github.com/erda-project/erda-infra/providers/pprof"
)

func main() {
	common.Run(&servicehub.RunOptions{
		ConfigFile: conf.MonitorStreamingConfigFilePath,
		Content:    conf.MonitorStreamingDefaultConfig,
	})
}
