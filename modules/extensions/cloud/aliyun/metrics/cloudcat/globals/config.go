// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package globals

import (
	"time"

	"github.com/erda-project/erda-infra/providers/kafka"
)

type Config struct {
	AccountReload     time.Duration        `file:"account_reload"`
	ProductListReload time.Duration        `file:"product_list_reload"`
	GatherWindow      time.Duration        `file:"gather_window"`
	MaxQPS            int                  `file:"max_qps" env:"CLOUDCAT_MAX_QPS" default:"1"`
	ReqLimit          uint64               `file:"req_limit" env:"CLOUDCAT_REQ_LIMIT" default:"3000"`
	ReqLimitDuration  time.Duration        `file:"req_limit_duration" env:"CLOUDCAT_REQ_LIMIT_DURATION" default:"1h"`
	ReqLimitTimeout   time.Duration        `file:"req_limit_timeout" env:"CLOUDCAT_REQ_LIMIT_TIMEOUT" default:"5m"`
	OrgIds            string               `file:"org_ids" env:"CLOUDCAT_IMPORT_ORG_IDS" desc:"list of orgs to import cloud telemetry"`
	ProductsCfg       string               `file:"products" env:"CLOUDCAT_PRODUCTS"`
	Output            kafka.ProducerConfig `file:"output"`
	Products          []string
}
