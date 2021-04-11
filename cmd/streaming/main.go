// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/erda-project/erda-infra/modcom"
	// providers and modules
	// _ "terminus.io/dice/monitor/modules/business/metrics/browser"
	// _ "terminus.io/dice/monitor/modules/business/trace/storage"
	// "terminus.io/dice/monitor/modules/common"
	_ "github.com/erda-project/erda/modules/monitor/core/logs/storage"
	// _ "terminus.io/dice/monitor/modules/domain/metrics/storage"

	_ "github.com/erda-project/erda-infra/providers/cassandra"
	// _ "terminus.io/dice/monitor/providers/elasticsearch"
	_ "github.com/erda-project/erda-infra/providers/kafka"
	// _ "terminus.io/dice/monitor/providers/mysql"
	// _ "terminus.io/dice/monitor/providers/telemetry"
)

func main() {
	modcom.RunWithCfgDir("conf/monitor/streaming", "streaming")
}
