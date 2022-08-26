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

package clickhouse

import (
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/creator"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
)

type config struct {
	QueryTimeout time.Duration `file:"query_timeout" default:"1m"`
}

type provider struct {
	Cfg       *config
	Log       logs.Logger
	Creator   creator.Interface   `autowired:"clickhouse.table.creator@entity" optional:"true"`
	Loader    loader.Interface    `autowired:"clickhouse.table.loader@entity"`
	Retention retention.Interface `autowired:"storage-retention-strategy@entity" optional:"true"`

	clickhouse clickhouse.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	svc := ctx.Service("clickhouse@eneity")
	if svc == nil {
		svc = ctx.Service("clickhouse")
	}
	if svc == nil {
		return fmt.Errorf("service clickhouse is required")
	}
	p.clickhouse = svc.(clickhouse.Interface)
	return nil
}

func init() {
	servicehub.Register("entity-storage-clickhouse", &servicehub.Spec{
		Services:             []string{"entity-storage-clickhouse-reader", "entity-storage-clickhouse-writer"},
		Dependencies:         []string{"clickhouse"},
		OptionalDependencies: []string{"clickhouse.table.creator"},
		ConfigFunc:           func() interface{} { return &config{} },
		Creator:              func() servicehub.Provider { return &provider{} },
	})
}
