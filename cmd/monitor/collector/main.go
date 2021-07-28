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

	// modules
	_ "github.com/erda-project/erda-infra/providers/health"
	_ "github.com/erda-project/erda-infra/providers/kafka"
	_ "github.com/erda-project/erda-infra/providers/pprof"

	// providers
	_ "github.com/erda-project/erda/modules/core/monitor/collector"
)

//go:generate sh -c "cd ${PROJ_PATH} && go generate -v -x github.com/erda-project/erda/modules/monitor/core/collector"
func main() {
	common.Run(&servicehub.RunOptions{
		ConfigFile: conf.MonitorCollectorConfigFilePath,
		Content:    conf.MonitorCollectorDefaultConfig,
	})
}
