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
	"github.com/erda-project/erda/modules/extensions/loghub"
	"github.com/erda-project/erda/pkg/common"
	"github.com/erda-project/erda/pkg/common/addon"

	// providers and modules
	_ "github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis"

	// // log export outputs
	_ "github.com/erda-project/erda/modules/extensions/loghub/exporter"
	_ "github.com/erda-project/erda/modules/extensions/loghub/exporter/output/elasticsearch"
	_ "github.com/erda-project/erda/modules/extensions/loghub/exporter/output/elasticsearch-proxy"
	_ "github.com/erda-project/erda/modules/extensions/loghub/exporter/output/stdout"
	_ "github.com/erda-project/erda/modules/extensions/loghub/exporter/output/udp"
	_ "github.com/erda-project/erda/modules/extensions/loghub/index/manager"

	// infra
	_ "github.com/erda-project/erda-infra/providers/health"
	_ "github.com/erda-project/erda-infra/providers/pprof"
)

func main() {
	common.RegisterInitializer(addon.OverrideEnvs)
	common.RegisterInitializer(loghub.Init)
	common.Run(&servicehub.RunOptions{})
}
